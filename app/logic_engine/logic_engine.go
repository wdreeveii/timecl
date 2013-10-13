package logic_engine

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"fmt"
	"github.com/robfig/revel"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type processor func(o *Object_t, objs map[int]*Object_t)

func (p processor) MarshalJSON() ([]byte, error) {
	return []byte("[]"), nil
}

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
	fmt.Printf("ID %4d  ", o["Id"])
	fmt.Printf("Type %10s  ", o["Type"])
	fmt.Printf("Source %3d  ", int(o["Source"].(float64)))
	fmt.Printf("Output %10f  ", o["Output"])
	fmt.Printf("Terminals: ")
	for _, val := range o["Terminals"].([]interface{}) {
		fmt.Printf("%d ", val)
	}
	fmt.Printf("\n")
}

func (o Object_t) Process(Objects map[int]*Object_t) {
}

func (o Object_t) AssignOutput(objs map[int]*Object_t, terminal int) {
	terms := o["Terminals"].([]interface{})
	obj := *objs[int(terms[terminal].(float64))]
	obj["NextOutput"] = o["Output"]
}

func (o Object_t) CheckTerminals(count int) bool {
	terms := o["Terminals"].([]interface{})
	if len(terms) < count {
		fmt.Println("Invalid ", o["Type"])
		return true
	}
	return false
}
func (o Object_t) GetTerminal(Objects map[int]*Object_t, term int) float64 {
	terms := o["Terminals"].([]interface{})
	theterm := int(terms[term].(float64))
	//fmt.Println("theterm:", theterm)
	obj := (*Objects[theterm])
	return obj["Output"].(float64)
}

func (o Object_t) GetProperty(name string) string {
	if o["PropertyCount"].(PropertyCount) <= 0 {
		return ""
	}
	for ii := PropertyCount(0); ii < o["PropertyCount"].(PropertyCount); ii++ {
		if o["PropertyNames"].(PropertyNames)[ii] == name {
			return o["PropertyValues"].(PropertyValues)[ii]
		}
	}
	fmt.Println("Unable to find property ", name, " for ", o["Type"])
	return ""
}

type Engine_t struct {
	mu              sync.Mutex
	Objects         map[int]*Object_t
	Index           int
	UpdateRate      float32
	SolveIterations int
}

func (e *Engine_t) Init() {
	e.UpdateRate = 10
	e.SolveIterations = 10
	e.Objects = make(map[int]*Object_t)
	e.LoadObjects()
	e.printObjects()
	go e.Start()
	go e.EngineClient()

}

func (e *Engine_t) Start() {
	e.Run()
}

func (e *Engine_t) Run() {
	for {
		e.mu.Lock()
		for ii := 0; ii < e.SolveIterations; ii++ {
			e.Process()
		}
		e.mu.Unlock()
		e.SetOutputs()
		e.printObjects()
		time.Sleep(time.Duration(1000/e.UpdateRate) * time.Millisecond)
	}
}

func (e *Engine_t) Process() {
	for _, val := range e.Objects {
		(*val)["process"].(processor)(val, e.Objects)
	}

	for _, val := range e.Objects {
		(*val)["Output"] = (*val)["NextOutput"]
	}
}

func (e *Engine_t) Save() {
	path, found := revel.Config.String("engine.savefile")
	if !found {
		return
	}
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	enc.Encode(e)
	err := ioutil.WriteFile(path, m.Bytes(), 0600)
	if err != nil {
		panic(err)
	}
}

func (e *Engine_t) LoadObjects() {
	path, found := revel.Config.String("engine.savefile")
	if !found {
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}
	n, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)
	err = dec.Decode(e)
	if err != nil {
		panic(err)
	}
	for k, _ := range e.Objects {
		obj := *e.Objects[k]
		obj["process"] = processors[obj["Type"].(string)]
	}
}

func (e *Engine_t) printObjects() {
	e.mu.Lock()
	for _, val := range e.Objects {
		val.Display()
	}
	e.mu.Unlock()
	// new line?
}

func (e *Engine_t) GetOutputs() {
	for { // range result list
		// if object not in list
		// return
		//e.Objects[id].Output = output_from_db
	}
}

func (e *Engine_t) SetOutputs() {
	// for each object set
}

type State_t struct {
	Id     int
	Output float64
}

func (e *Engine_t) GetStates() []State_t {
	e.mu.Lock()
	var states []State_t
	for _, val := range e.Objects {
		states = append(states, State_t{Id: (*val)["Id"].(int), Output: (*val)["Output"].(float64)})
	}
	e.mu.Unlock()
	return states
}

func (e *Engine_t) HookObject(id int, source int) {
	e.mu.Lock()
	(*e.Objects[id])["Source"] = source
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) UnhookObject(id int) {
	e.mu.Lock()
	(*e.Objects[id])["Source"] = -1
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) ListObjects() []Object_t {
	e.mu.Lock()
	objs := make([]Object_t, 0, len(e.Objects))
	for _, val := range e.Objects {
		objs = append(objs, *val)
	}
	e.mu.Unlock()
	return objs
}

func (e *Engine_t) AddObject(obj Object_t) {
	e.mu.Lock()
	var id int
	id = int(obj["Id"].(float64))
	obj["Id"] = id
	obj["process"] = processors[obj["Type"].(string)]
	e.Objects[id] = &obj
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) DeleteObject(id int) {
	e.mu.Lock()
	delete(e.Objects, id)
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) MoveObject(id, x_pos, y_pos int) {
	fmt.Println("Move: ", id, x_pos, y_pos)
	e.mu.Lock()
	(*e.Objects[id])["Xpos"] = x_pos
	(*e.Objects[id])["Ypos"] = y_pos
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) SetGuides(id int, guide int) {
	e.mu.Lock()
	(*e.Objects[id])["Terminals"] = append((*e.Objects[id])["Terminals"].(Terminals), guide)
	e.Save()
	e.mu.Unlock()
}

func (e *Engine_t) SetOutput(id int, output float64) {
	e.mu.Lock()
	fmt.Println("Setting output...", output)
	(*e.Objects[id])["Output"] = output
	e.Save()
	e.mu.Unlock()
}
func (e *Engine_t) SetProperties(id int, property_count int,
	property_names []string, property_types []string, property_values []string) {
	e.mu.Lock()
	(*e.Objects[id])["PropertyNames"] = property_names
	(*e.Objects[id])["PropertyTypes"] = property_types
	(*e.Objects[id])["PropertyValues"] = property_values
	e.Save()
	e.mu.Unlock()
}
func (e *Engine_t) EngineClient() {
	subscription := e.Subscribe()
	defer subscription.Cancel()

	for {
		event := <-subscription.New
		switch {
		case event.Type == "add":
			fmt.Println("add")
			obj := event.Data.(map[string]interface{})
			e.AddObject(obj)
		case event.Type == "edit":
			fmt.Println("edit recv")
		}

	}
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

func PublishStateChange(id int, state map[string]interface{}) {
	change_event := StateChange{Id: id, State: state}
	publish <- newEvent("edit", change_event)
}

func (e *Engine_t) Publish(event Event) {
	publish <- event
}

func testPublish() {
	for {
		new_state := map[string]interface{}{
			"a": 123,
			"b": "hello",
		}
		fmt.Println("create state:", new_state)
		PublishStateChange(0, new_state)
		time.Sleep(5 * time.Second)
	}
}

const archiveSize = 10

var (
	subscribe   = make(chan (chan<- Subscription), 10)
	unsubscribe = make(chan (<-chan Event), 10)
	publish     = make(chan Event, 10)
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
			subscriber := make(chan Event, 10)
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

func Init() {
	fmt.Println("engine start")
}

func init() {
	revel.OnAppStart(Init)
	go engine_pub_sub()
	go testPublish()
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
