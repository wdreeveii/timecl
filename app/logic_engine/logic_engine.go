package logic_engine

import (
	"container/list"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime/pprof"
	"sort"
	"time"
	"timecl/app/logger"
	"timecl/app/network_manager"
)

var DEBUG = log.New(ioutil.Discard, "LogicEngine ", log.Ldate|log.Ltime)

type addObjRequest struct {
	Type Type_t
	X    Dim_t
	Y    Dim_t
}
type Engine_t struct {
	Objects         ObjectList
	Index           Id_t
	UpdateRate      float32
	SolveIterations int
	list_objs       chan chan []Object_t
	delete_obj      chan Id_t
	unhook_obj      chan Id_t
	add_obj         chan addObjRequest
	DataFile        string
	subscribe       chan (chan<- Subscription)
	unsubscribe     chan (<-chan Event)
	stopRouter      chan (chan<- bool)
	stopEngine      chan (chan<- bool)
	publish         chan Event
}

func Init(dataFile string) *Engine_t {
	DEBUG.Println("Logic Engine Start")
	var e Engine_t
	e.subscribe = make(chan (chan<- Subscription), 10)
	e.unsubscribe = make(chan (<-chan Event), 10)
	e.publish = make(chan Event, 100)
	e.stopRouter = make(chan (chan<- bool))
	go e.engine_pub_sub()

	e.list_objs = make(chan chan []Object_t)
	e.delete_obj = make(chan Id_t)
	e.unhook_obj = make(chan Id_t)
	e.add_obj = make(chan addObjRequest)
	e.UpdateRate = 10
	e.SolveIterations = 50
	e.Objects = make(ObjectList)
	e.DataFile = dataFile
	e.LoadObjects()
	e.DataFile = dataFile
	e.stopEngine = make(chan (chan<- bool))
	go e.run()
	return &e
}

func (e *Engine_t) Stop() {
	router_stop := make(chan bool)
	engine_stop := make(chan bool)
	e.stopRouter <- router_stop
	e.stopEngine <- engine_stop

	<-router_stop
	<-engine_stop
}

func (e *Engine_t) deleteObject(id Id_t) {
	e.delete_obj <- id
}

func (e *Engine_t) _deleteObject(id Id_t) {
	delete(e.Objects, id)
	e.Save()
	e.Publish(Event{Type: "del", Data: id})
}

func (e *Engine_t) unhookObject(id Id_t) {
	e.unhook_obj <- id
}

func (e *Engine_t) _unhookSingleObject(id Id_t) {
	fmt.Println("unhook obj", id)
	obj, exists := e.Objects[id]
	if exists {
		obj.Source = -1
		e.Objects[id] = obj
	}
	newstate := make(map[string]interface{})
	newstate["Source"] = -1
	change := StateChange{Id: id, State: newstate}
	e.Publish(Event{Type: "edit", Data: change})
}

func (e *Engine_t) _unhookBase(id Id_t) {
	e._unhookConnectedObjects(id)
	e._unhookSingleObject(id)
}

func (e *Engine_t) _unhookObject(id Id_t) {
	fmt.Println("unhook obj", id)
	e._unhookBase(id)
	for _, v := range e.Objects[id].Terminals {
		fmt.Println("unhooking terminal", v)
		e._unhookBase(v)

	}
}

// change to reclaim unused indexes in the future
func (e *Engine_t) getUnusedIndex() Id_t {
	var max = Id_t(0)
	for k, _ := range e.Objects {
		if k > max {
			max = k
		}
	}
	max++
	e.Objects[max] = &Object_t{}
	return max
}

func (e *Engine_t) add_input_terminal(pos Dim_t, parent *Object_t) {
	new_id := e.getUnusedIndex()
	new_obj := e.init_obj(new_id, Type_t("guide"),
		parent.Xpos-10,
		parent.Ypos+parent.Ysize/2-10/2+pos*(10+2))
	new_obj.Dir = DirLeft
	new_obj.Attached = 1
	e.Objects[new_id] = &new_obj
	e.Publish(Event{Type: "add", Data: new_obj})

	parent.Terminals = append(parent.Terminals, new_id)
}

func (e *Engine_t) add_output_terminal(pos Dim_t, parent *Object_t) {
	new_id := e.getUnusedIndex()
	new_obj := e.init_obj(new_id, Type_t("guide"),
		parent.Xpos+parent.Xsize,
		parent.Ypos+parent.Ysize/2-10/2+pos*(10+2))
	new_obj.Dir = DirRight
	new_obj.Attached = 1
	e.Objects[new_id] = &new_obj
	e.Publish(Event{Type: "add", Data: new_obj})

	parent.Terminals = append(parent.Terminals, new_id)
}
func (e *Engine_t) init_obj(new_id Id_t, obj_type Type_t, x Dim_t, y Dim_t) Object_t {
	o := inits[obj_type]()
	o.Type = obj_type
	o.Xpos = Dim_t(x)
	o.Ypos = Dim_t(y)
	o.Id = new_id
	return o
}

func (e *Engine_t) addObject(obj_type string, x float64, y float64) {
	var request = addObjRequest{Type: Type_t(obj_type), X: Dim_t(x), Y: Dim_t(y)}
	e.add_obj <- request
}

func (e *Engine_t) _addObject(request addObjRequest) {
	new_id := e.getUnusedIndex()
	new_obj := e.init_obj(new_id, request.Type, request.X, request.Y)

	for ii := 0; ii < new_obj.inputTermCount; ii++ {
		if new_obj.inputTermCount == 1 {
			e.add_input_terminal(Dim_t(0), &new_obj)
		} else if new_obj.inputTermCount == 2 {
			e.add_input_terminal(Dim_t(2*ii-1), &new_obj)
		}
	}
	for ii := 0; ii < new_obj.outputTermCount; ii++ {
		if new_obj.outputTermCount == 1 {
			e.add_output_terminal(Dim_t(0), &new_obj)
		} else if new_obj.outputTermCount == 2 {
			e.add_output_terminal(Dim_t(2*ii-1), &new_obj)
		}
	}
	e.Objects[new_id] = &new_obj
	e.Publish(Event{Type: "add", Data: new_obj})
}

func (e *Engine_t) _unhookConnectedObjects(id Id_t) {
	for k, v := range e.Objects {
		if v.Source == id {
			e._unhookSingleObject(k)
		}
	}
}

func (e *Engine_t) _unhookForDelete(id Id_t) {
	e._unhookConnectedObjects(id)
	for _, v := range e.Objects[id].Terminals {
		e._unhookConnectedObjects(v)
	}
}

func (e *Engine_t) editObject(new_states StateChange) {
	save_obj := false
	id := new_states.Id
	obj, ok := e.Objects[id]
	if ok {
		for k, v := range new_states.State {
			ps := reflect.ValueOf(obj)
			s := ps.Elem()
			if s.Kind() == reflect.Struct {
				f := s.FieldByName(k)
				if f.IsValid() {
					if f.CanSet() {
						fmt.Println("trying to set", k, v)
						f.Set(reflect.ValueOf(v).Convert(f.Type()))
						save_obj = true
					}
				}
			}
			fmt.Println("key val", k, v, save_obj)
		}
		if save_obj {
			sanitize(obj)
			e.Save()
		}
	} else {
		logger.PublishOneError(fmt.Errorf("Edit: Unknown object %v", id))
	}
}

func (e *Engine_t) store_outputs() (outputs map[Id_t]Value_t) {
	outputs = make(map[Id_t]Value_t, len(e.Objects))
	for k, val := range e.Objects {
		outputs[k] = val.Output
		otype := val.Type
		switch {
		case otype == "binput",
			otype == "ainput":
			iport_uri, err := val.GetProperty("port")
			if err != nil {
				logger.PublishOneError(err)
			} else {
				if port_uri, ok := iport_uri.(network_manager.PortURI); ok {
					newvalue, err := network_manager.Get(port_uri)
					if err == nil {
						val.SignalGood = true
						val.PortValue = Value_t(newvalue) // propagate this type into network_manager
					} else {
						val.SignalGood = false
					}
				}
			}
		}
	}
	return
}

func (e *Engine_t) publish_output_changes(outputs map[Id_t]Value_t) {
	var state_changes []StateChange
	var output_changes []network_manager.PortChange

	for k, val := range e.Objects {
		if outputs[k] != val.Output {
			newstate := make(map[string]interface{})
			newstate["Output"] = val.Output
			change := StateChange{Id: k, State: newstate}
			state_changes = append(state_changes, change)
			otype := val.Type
			switch {
			case otype == "boutput",
				otype == "aoutput":
				iport_uri, err := val.GetProperty("port")
				if err != nil {
					logger.PublishOneError(err)
				} else {
					if port_uri, ok := iport_uri.(network_manager.PortURI); ok {
						output_changes = append(output_changes, network_manager.PortChange{URI: port_uri, Value: network_manager.Value_t(val.Output)})
					}
				}
			}
		}
	}
	if len(output_changes) > 0 {
		network_manager.PublishSetEvents(output_changes)
	}
	if len(state_changes) > 0 {
		e.PublishMultipleStateChanges(state_changes)
	}
}

func (e *Engine_t) run() {
	var profile_timeout <-chan time.Time
	path, found := revel.Config.String("engine.profilepath")
	if found {
		f, err := os.Create(path)
		if err != nil {
			panic("Can't create profile.")
		}
		profile_timeout = time.After(30 * time.Minute)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var calc_world <-chan time.Time
	calc_world = time.After(time.Duration(1000/e.UpdateRate) * time.Millisecond)
	engine_subscription := e.Subscribe()
	defer e.CancelSubscription(engine_subscription)

	network_subscription := network_manager.Subscribe()
	defer network_subscription.Cancel()
	for {
		select {
		case receiver := <-e.list_objs:
			objs := make([]Object_t, 0, len(e.Objects))
			for _, val := range e.Objects {
				objs = append(objs, *val)
			}
			receiver <- objs
		case resp := <-e.stopEngine:
			resp <- true
			return
		case <-profile_timeout:
			fmt.Println("Profile Done. Exiting...")
			return
		case obj_id := <-e.delete_obj:
			obj, exists := e.Objects[obj_id]
			if exists {
				e._unhookForDelete(obj_id)
				fmt.Println("terminals", obj.Terminals)
				for _, v := range obj.Terminals {
					e._deleteObject(v)
				}
				e._deleteObject(obj_id)
				e.Save()
			}
		case obj_id := <-e.unhook_obj:
			_, exists := e.Objects[obj_id]
			if exists {
				e._unhookObject(obj_id)
				e.Save()
			}
		case request := <-e.add_obj:
			e._addObject(request)
			e.Save()
		case event, ok := <-engine_subscription.New:
			if !ok {
				return
			}
			switch {
			case event.Type == "edit":
				var s StateChange
				switch v := event.Data.(type) {
				case map[string]interface{}:
					s = StateChange{Id: Id_t(v["Id"].(float64)), State: v["State"].(map[string]interface{})}
				case StateChange:
					s = v
				}
				e.editObject(s)
			default:
				fmt.Println("unrecognized object:", event.Type, event.Data)
			}
		case event := <-network_subscription.New:
			DEBUG.Println("Engine Event")
			DEBUG.Println(event.NetworkID)
			DEBUG.Println(event.Type)
			switch {
			case event.Type == "port_change":
				// send port list to the clients
			case event.Type == "state_change":
				// update the engine with the new state
			}
		case <-calc_world:
			calc_world = time.After(time.Duration(1000/e.UpdateRate) * time.Millisecond)
			var outputs = e.store_outputs()

			var errDedup = make(map[string]*logger.ErrInfo)
			for ii := 0; ii < e.SolveIterations; ii++ {
				for _, val := range e.Objects {
					process := val.process
					if process != nil {
						err := process(val, e.Objects, ii)
						if err != nil {
							errkey := fmt.Sprintf("Process Object: %v", err)
							_, ok := errDedup[errkey]
							if ok {
								errDedup[errkey].Count = 1
								errDedup[errkey].Time = time.Now()
							} else {
								errDedup[errkey] = &logger.ErrInfo{Count: 1, Time: time.Now(), First: time.Now()}
							}
						}
					}
				}

				for _, val := range e.Objects {
					val.Output = val.NextOutput
				}
			}
			var errList = make(logger.ErrorSlice, 0, len(errDedup))
			for k, v := range errDedup {
				errList = append(errList, &logger.ErrorListElement{Error: k, Count: v.Count, Time: v.Time, First: v.First})
			}
			if len(errList) > 0 {
				sort.Sort(errList)
				logger.Publish(logger.Event{Type: "errors", Data: errList})
			}
			e.publish_output_changes(outputs)
		}
	}
}

func (e *Engine_t) ListObjects() Event {
	var objs []Object_t
	if e.list_objs != nil {
		var res = make(chan []Object_t)
		e.list_objs <- res
		objs = <-res
	}
	event := newEvent("init", objs)
	return event
}

func (e *Engine_t) ListPorts() Event {
	var ports = network_manager.ListPorts()
	event := newEvent("init_ports", ports)
	return event
}

type Subscription struct {
	New <-chan Event
}

func (e *Engine_t) CancelSubscription(s Subscription) {
	e.unsubscribe <- s.New
	drain(s.New)
}

func (e *Engine_t) Subscribe() Subscription {
	resp := make(chan Subscription)
	e.subscribe <- resp
	return <-resp
}

type Event struct {
	Type string
	Data EventArgument
}

func newEvent(typ string, data EventArgument) Event {
	return Event{typ, data}
}

type StateChange struct {
	Id    Id_t
	State map[string]interface{}
}

type EventArgument interface {
}

func (e *Engine_t) PublishMultipleStateChanges(updates []StateChange) {
	e.publish <- newEvent("edit_many", updates)
}

func (e *Engine_t) Publish(event Event) {
	e.publish <- event
}

func (e *Engine_t) engine_pub_sub() {
	subscribers := list.New()

	for {
		select {
		case ch := <-e.subscribe:
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}
		case event := <-e.publish:
			switch event.Type {
			case "errors":
				panic("Unused event")
			case "delete_object":
				obj_id, ok := event.Data.(float64)
				if ok {
					e.deleteObject(Id_t(obj_id))
				}
			case "add_object":
				objinfo, ok := event.Data.(map[string]interface{})
				if ok {
					e.addObject(objinfo["Type"].(string), objinfo["X"].(float64), objinfo["Y"].(float64))
				}
			case "unhook":
				obj_id, ok := event.Data.(float64)
				if ok {
					e.unhookObject(Id_t(obj_id))
				}
			default:
				for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
					ch.Value.(chan Event) <- event
				}
			}
		case unsub := <-e.unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		case resp := <-e.stopRouter:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				close(ch.Value.(chan Event))
			}
			resp <- true
			return
		}
	}
}

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
