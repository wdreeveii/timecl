package logic_engine

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"time"
	"timecl/app/logger"
	"timecl/app/network_manager"
)

var DEBUG = log.New(ioutil.Discard, "LogicEngine ", log.Ldate|log.Ltime)

type Engine_t struct {
	Objects         map[int]Object_t
	Index           int
	UpdateRate      float32
	SolveIterations int
	list_objs       chan chan []Object_t
}

func (e *Engine_t) Init() {
	DEBUG.Println("Logic Engine Start")
	e.UpdateRate = 10
	e.SolveIterations = 50
	e.Objects = make(map[int]Object_t)
	e.LoadObjects()
	go e.run()
}

func (e *Engine_t) addObject(obj Object_t) {
	obj["process"] = processors[stringify(obj["Type"])]
	sanitize(obj)
	e.Objects[intify(obj["Id"])] = obj
	e.Save()
}

func (e *Engine_t) deleteObject(id int) {
	delete(e.Objects, id)
	e.Save()
}

func (e *Engine_t) editObject(new_states StateChange) {
	save_obj := false
	id := new_states.Id
	obj, ok := e.Objects[id]
	if ok {
		for k, v := range new_states.State {
			switch val := v.(type) {
			case []interface{}:
				slice := obj[k].([]interface{})
				for k2, v2 := range val {
					if slice[k2] != v2 {
						slice[k2] = v2
						save_obj = true
					}
				}
			case interface{}:
				if obj[k] != v {
					obj[k] = v
					save_obj = true
				}
			}
		}
		if save_obj {
			sanitize(obj)
			e.Save()
		}
	} else {
		logger.PublishOneError(errors.New("Edit: Unknown object"))
	}
}

func (e *Engine_t) store_outputs() (outputs map[int]float64) {
	outputs = make(map[int]float64, len(e.Objects))
	for k, val := range e.Objects {
		outputs[k] = val["Output"].(float64)
		otype := val["Type"]
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
						val["PortValue"] = newvalue
					} else {
						delete(val, "PortValue")
					}
				}
			}
		}
	}
	return
}

func (e *Engine_t) publish_output_changes(outputs map[int]float64) {
	var state_changes []StateChange
	var output_changes []network_manager.PortChange

	for k, val := range e.Objects {
		if outputs[k] != val["Output"] {
			newstate := make(map[string]interface{})
			newstate["Output"] = val["Output"].(float64)
			change := StateChange{Id: k, State: newstate}
			state_changes = append(state_changes, change)
			otype := val["Type"]
			switch {
			case otype == "boutput",
				otype == "aoutput":
				iport_uri, err := val.GetProperty("port")
				if err != nil {
					logger.PublishOneError(err)
				} else {
					if port_uri, ok := iport_uri.(network_manager.PortURI); ok {
						output_changes = append(output_changes, network_manager.PortChange{URI: port_uri, Value: val["Output"].(float64)})
					}
				}
			}
		}
	}
	if len(output_changes) > 0 {
		network_manager.PublishSetEvents(output_changes)
	}
	if len(state_changes) > 0 {
		PublishMultipleStateChanges(state_changes)
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
	defer engine_subscription.Cancel()

	network_subscription := network_manager.Subscribe()
	defer network_subscription.Cancel()
	e.list_objs = make(chan chan []Object_t)
	for {
		select {
		case receiver := <-e.list_objs:
			var obj Object_t
			objs := make([]Object_t, 0, len(e.Objects))
			for _, val := range e.Objects {
				obj = make(Object_t)
				for key, prop := range val {
					obj[key] = prop
				}
				objs = append(objs, obj)
			}
			receiver <- objs
		case <-profile_timeout:
			fmt.Println("Profile Done. Exiting...")
			return
		case event := <-engine_subscription.New:
			switch {
			case event.Type == "add":
				obj := event.Data.(map[string]interface{})
				e.addObject(obj)
			case event.Type == "edit_many":
				state_changes := event.Data.([]StateChange)
				for _, v := range state_changes {
					e.editObject(v)
				}
			case event.Type == "edit":
				var s StateChange
				switch v := event.Data.(type) {
				case map[string]interface{}:
					s = StateChange{Id: int(v["Id"].(float64)), State: v["State"].(map[string]interface{})}
				case StateChange:
					s = v
				}
				e.editObject(s)
			case event.Type == "del":
				var id int
				data := event.Data.(map[string]interface{})
				id = intify(data["Id"])
				e.deleteObject(id)
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
					process := val["process"].(processor)
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
					val["Output"] = val["NextOutput"]
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

func (s Subscription) Cancel() {
	unsubscribe <- s.New
	drain(s.New)
}

func (e *Engine_t) Subscribe() Subscription {
	resp := make(chan Subscription)
	subscribe <- resp
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
	Id    int
	State map[string]interface{}
}

type EventArgument interface {
}

func PublishMultipleStateChanges(updates []StateChange) {
	publish <- newEvent("edit_many", updates)
}

func (e *Engine_t) Publish(event Event) {
	publish <- event
}

const archiveSize = 10

var (
	subscribe   = make(chan (chan<- Subscription), 10)
	unsubscribe = make(chan (<-chan Event), 10)
	publish     = make(chan Event, 100)
)

// This function loops forever, handling the chat room pubsub
func engine_pub_sub() {
	subscribers := list.New()
	var errDedup = make(map[string]*logger.ErrInfo)
	errorTicker := time.Tick(1 * time.Second)
	logger_subscription := logger.Subscribe()

	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}
		case <-errorTicker:
			var errors logger.ErrorSlice
			for k, v := range errDedup {
				errors = append(errors, &logger.ErrorListElement{Error: k,
					Count: v.Count,
					Time:  v.Time,
					First: v.First})
			}
			errDedup = make(map[string]*logger.ErrInfo)
			if len(errors) > 0 {
				sort.Sort(errors)
				for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
					ch.Value.(chan Event) <- Event{Type: "error_list", Data: errors}
				}
			}
		case event := <-logger_subscription.New:
			switch event.Type {
			case "errors":
				errlist, ok := event.Data.(logger.ErrorSlice)
				if ok {
					for _, v := range errlist {
						cachederr, ok := errDedup[v.Error]
						if ok {
							if v.Time.After(cachederr.Time) {
								errDedup[v.Error].Time = v.Time
							}
							errDedup[v.Error].Count += v.Count
						} else {
							errDedup[v.Error] = &logger.ErrInfo{Count: v.Count,
								Time:  v.Time,
								First: v.First}
						}
					}
				}
			}
		case event := <-publish:
			switch event.Type {
			case "errors":
				panic("Unused event")
			default:
				for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
					ch.Value.(chan Event) <- event
				}
			}
		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		}
	}
}

func init() {
	go engine_pub_sub()
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
