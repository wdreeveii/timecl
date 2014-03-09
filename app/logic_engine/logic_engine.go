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
	"time"
	"timecl/app/network_manager"
)

var LOG = log.New(os.Stderr, "LogicEngine ", log.Ldate|log.Ltime)
var DEBUG = log.New(ioutil.Discard, "LogicEngine ", log.Ldate|log.Ltime)

type processor func(o *Object_t, objs map[int]*Object_t, iteration int)

type Id int
type Type string
type Xpos int
type Ypos int
type Xsize int
type ysize int
type Output float64
type NextOutput float64
type TermList string
type Terminals []int
type Source int
type PropertyCount int
type PropertyNames []string
type PropertyValues []string
type PropertyTypes []string
type Attached int
type Dir int

type Object_t map[string]interface{}

/*type Object_t struct {
	Id             int
	Type           string
	Xpos           int
	Ypos           int
	Xsize          int
	Ysize          int
	Output         float64
	NextOutput     float64
	TermList       string
	Terminals      []int
	Source         int
	PropertyCount  int
	PropertyNames  []string
	PropertyValues []string
	PropertyTypes  []string
	Attached       int
	Dir            int
	process        processor
}*/

func (o Object_t) Display() {
	var output string
	output += fmt.Sprintf("ID %4d  ", o["Id"])
	output += fmt.Sprintf("Type %10s  ", o["Type"])
	output += fmt.Sprintf("Source %3d  ", int(o["Source"].(int)))
	output += fmt.Sprintf("Output %10f  ", o["Output"])
	output += fmt.Sprintf("Terminals: ")
	for _, val := range o["Terminals"].([]interface{}) {
		output += fmt.Sprintf("%d ", int(val.(float64)))
	}
	LOG.Println(output)
}

func (o Object_t) Process(Objects map[int]*Object_t) {
}

func (o Object_t) AssignOutput(objs map[int]*Object_t, terminal int) {
	iterms, ok := o["Terminals"]
	if !ok {
		LOG.Println("No terminal list.")
		return
	}
	terms, ok := iterms.([]interface{})
	if !ok {
		LOG.Println("Terminal list of unknown type.")
		return
	}

	terminal64, ok := terms[terminal].(float64)
	if !ok {
		LOG.Println("Terminal conversion error.")
		return
	}
	obj, ok := objs[int(terminal64)]
	if !ok {
		LOG.Println("The specified object does not exist.")
		return
	}
	output, ok := o["Output"]
	if !ok {
		LOG.Println("No output.")
		return
	}
	output64, ok := output.(float64)
	if !ok {
		LOG.Println("Output of unknown type.")
		return
	}
	(*obj)["NextOutput"] = output64
}

func (o Object_t) CheckTerminals(count int) bool {
	iterms, ok := o["Terminals"]
	if !ok {
		LOG.Println("No terminal list.")
		return true
	}
	terms, ok := iterms.([]interface{})
	if !ok {
		LOG.Println("Terminal list of unknown type.")
		return true
	}
	if len(terms) < count {
		LOG.Println("Invalid Terminals for obj type:", o["Type"])
		return true
	}
	return false
}

func (o Object_t) GetTerminal(Objects map[int]*Object_t, term int) (float64, error) {
	iterms, ok := o["Terminals"]
	if !ok {
		return 0, errors.New("No terminal list.")
	}
	terms, ok := iterms.([]interface{})
	if !ok {
		return 0, errors.New("Terminals list of unknown type.")
	}
	terminal64, ok := terms[term].(float64)
	if !ok {
		return 0, errors.New("Terminal conversion error.")
	}
	theterm := int(terminal64)
	obj, ok := Objects[theterm]
	if !ok {
		return 0, errors.New("Specified object does not exist.")
	}
	output, ok := (*obj)["Output"]
	if !ok {
		return 0, errors.New("No output.")
	}
	output64, ok := output.(float64)
	if !ok {
		return 0, errors.New("Output of unknown type.")
	}
	return output64, nil
}

func (o Object_t) GetProperty(name string) interface{} {
	PCount := o["PropertyCount"].(int)
	if PCount <= 0 {
		return nil
	}
	names := o["PropertyNames"].([]interface{})
	for ii := 0; ii < PCount; ii++ {
		if stringify(names[ii]) == name {
			return o["PropertyValues"].([]interface{})[ii]
		}
	}
	LOG.Println("Unable to find property ", name, " for ", o["Type"])
	return nil
}

type Engine_t struct {
	Objects         map[int]*Object_t
	Index           int
	UpdateRate      float32
	SolveIterations int
	list_objs       chan chan []Object_t
}

func (e *Engine_t) Init() {
	LOG.Println("Logic Engine Start")
	e.UpdateRate = 10
	e.SolveIterations = 50
	e.Objects = make(map[int]*Object_t)
	e.LoadObjects()
	go e.Run()
}

func (e *Engine_t) AddObject(obj Object_t) {
	obj["process"] = processors[stringify(obj["Type"])]
	sanitize(&obj)
	e.Objects[intify(obj["Id"])] = &obj
	e.Save()
}

func (e *Engine_t) DeleteObject(id int) {
	delete(e.Objects, id)
	e.Save()
}

func (e *Engine_t) EditObject(new_states StateChange) {
	save_obj := false
	id := new_states.Id
	obj, ok := e.Objects[id]
	if ok {
		for k, v := range new_states.State {
			switch val := v.(type) {
			case []interface{}:
				slice := (*obj)[k].([]interface{})
				for k2, v2 := range val {
					if slice[k2] != v2 {
						slice[k2] = v2
						save_obj = true
					}
				}
			case interface{}:
				if (*obj)[k] != v {
					(*obj)[k] = v
					save_obj = true
				}
			}
		}
		if save_obj {
			sanitize(obj)
			e.Save()
		}
	} else {
		LOG.Println("Edit: Unknown object")
	}
}

func (e *Engine_t) store_outputs() (outputs map[int]float64) {
	outputs = make(map[int]float64, len(e.Objects))
	for k, val := range e.Objects {
		outputs[k] = (*val)["Output"].(float64)
		otype := (*val)["Type"]
		switch {
		case otype == "binput",
			otype == "ainput":
			if port_uri, ok := (*val).GetProperty("port").(network_manager.PortURI); ok {
				newvalue, err := network_manager.Get(port_uri)
				if err == nil {
					(*val)["PortValue"] = newvalue
				} else {
					delete((*val), "PortValue")
					//LOG.Println("Problem getting port value:", err)
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
		if outputs[k] != (*val)["Output"] {
			newstate := make(map[string]interface{})
			newstate["Output"] = (*val)["Output"].(float64)
			change := StateChange{Id: k, State: newstate}
			state_changes = append(state_changes, change)
			otype := (*val)["Type"]
			switch {
			case otype == "boutput",
				otype == "aoutput":
				if port_uri, ok := (*val).GetProperty("port").(network_manager.PortURI); ok {
					output_changes = append(output_changes, network_manager.PortChange{URI: port_uri, Value: (*val)["Output"].(float64)})
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

func (e *Engine_t) Run() {
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
			objs := make([]Object_t, 0, len(e.Objects))
			for _, val := range e.Objects {
				objs = append(objs, *val)
			}
			receiver <- objs
		case <-profile_timeout:
			fmt.Println("Profile Done. Exiting...")
			return
		case event := <-engine_subscription.New:
			switch {
			case event.Type == "add":
				obj := event.Data.(map[string]interface{})
				e.AddObject(obj)
			case event.Type == "edit_many":
				state_changes := event.Data.([]StateChange)
				for _, v := range state_changes {
					e.EditObject(v)
				}
			case event.Type == "edit":
				var s StateChange
				switch v := event.Data.(type) {
				case map[string]interface{}:
					s = StateChange{Id: int(v["Id"].(float64)), State: v["State"].(map[string]interface{})}
				case StateChange:
					s = v
				}
				e.EditObject(s)
			case event.Type == "del":
				var id int
				data := event.Data.(map[string]interface{})
				id = intify(data["Id"])
				e.DeleteObject(id)
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

			for ii := 0; ii < e.SolveIterations; ii++ {
				for _, val := range e.Objects {
					process := (*val)["process"].(processor)
					if process != nil {
						process(val, e.Objects, ii)
					}
				}

				for _, val := range e.Objects {
					(*val)["Output"] = (*val)["NextOutput"]
				}
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
	Archive []Event
	New     <-chan Event
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
	Type      string
	Data      EventArgument
	Timestamp int
}

func newEvent(typ string, data EventArgument) Event {
	return Event{typ, data, int(time.Now().Unix())}
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
	archive := list.New()
	subscribers := list.New()

	for {
		select {
		case ch := <-subscribe:
			var events []Event
			for e := archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{events, subscriber}

		case event := <-publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
			if archive.Len() >= archiveSize {
				archive.Remove(archive.Front())
			}
			archive.PushBack(event)

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
