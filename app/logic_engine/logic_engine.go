package logic_engine

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"fmt"
	"github.com/robfig/revel"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

type processor func(o *Object_t, objs map[int]*Object_t)

func (p processor) MarshalJSON() ([]byte, error) {
	return []byte("[]"), nil
}

func (p *processor) GobEncode() ([]byte, error) {
	return []byte(""), nil
}
func (p *processor) GobDecode([]byte) error {
	return nil
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
	fmt.Printf("Source %3d  ", int(o["Source"].(int)))
	fmt.Printf("Output %10f  ", o["Output"])
	fmt.Printf("Terminals: ")
	for _, val := range o["Terminals"].([]interface{}) {
		fmt.Printf("%d ", int(val.(float64)))
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

func (o Object_t) GetProperty(name string) interface{} {
	PCount := o["PropertyCount"].(int)
	if PCount <= 0 {
		return nil
	}
	for ii := 0; ii < PCount; ii++ {
		if stringify(o["PropertyNames"].([]interface{})[ii]) == name {
			return o["PropertyValues"].([]interface{})[ii]
		}
	}
	fmt.Println("Unable to find property ", name, " for ", o["Type"])
	return nil
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
		outputs := make(map[int]float64, len(e.Objects))
		for k, val := range e.Objects {
			outputs[k] = (*val)["Output"].(float64)
		}
		for ii := 0; ii < e.SolveIterations; ii++ {
			e.Process()
		}
		for k, val := range e.Objects {
			if outputs[k] != (*val)["Output"] {
				newstate := make(map[string]interface{})
				newstate["Output"] = (*val)["Output"].(float64)
				PublishStateChange(k, newstate)
			}
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
	tmp := make([]interface{}, 0)
	gob.Register(tmp)
	var p processor
	gob.Register(p)
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(e)
	if err != nil {
		fmt.Println("Encoding:", err)
	}
	err = ioutil.WriteFile(path, m.Bytes(), 0600)
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
		fmt.Println(err)
		return
	}
	tmp := make([]interface{}, 0)
	gob.Register(tmp)
	var proc processor
	gob.Register(proc)
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)

	err = dec.Decode(e)
	if err != nil {
		fmt.Println(err)
		return
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

func (e *Engine_t) ListObjects() Event {
	e.mu.Lock()
	objs := make([]Object_t, 0, len(e.Objects))
	for _, val := range e.Objects {
		objs = append(objs, *val)
	}
	e.mu.Unlock()
	event := newEvent("init", objs)
	return event
}

func floatify(in interface{}) float64 {
	var result float64
	var err error
	switch v := in.(type) {
	case string:
		result, err = strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Println(err)
		}
	case float64:
		result = v
	case int:
		result = float64(v)
	}
	return result
}

func intify(in interface{}) int {
	var result int
	switch v := in.(type) {
	case float64:
		result = int(v)
	case int:
		result = v
	}
	return result
}

func stringify(in interface{}) string {
	var result string
	switch v := in.(type) {
	case float64:
		result = strconv.FormatFloat(v, 'f', 3, 64)
	case string:
		result = v
	case int:
		result = strconv.FormatInt(int64(v), 10)
	}
	return result
}

func sanitize(obj Object_t) Object_t {
	var source int
	source = intify(obj["Source"])
	obj["Source"] = source

	var PCount int
	PCount = intify(obj["PropertyCount"])
	obj["PropertyCount"] = PCount

	PNames := make([]interface{}, 0)
	for _, v := range obj["PropertyNames"].([]interface{}) {
		PNames = append(PNames, stringify(v))
	}
	obj["PropertyNames"] = PNames

	PTypes := make([]interface{}, 0)
	for _, v := range obj["PropertyTypes"].([]interface{}) {
		PTypes = append(PTypes, stringify(v))
	}
	obj["PropertyTypes"] = PTypes

	PValues := make([]interface{}, 0)
	for k, v := range obj["PropertyValues"].([]interface{}) {
		switch {
		case PTypes[k] == "float":
			PValues = append(PValues, floatify(v))
		case PTypes[k] == "string" || PTypes[k] == "time":
			PValues = append(PValues, stringify(v))
		case PTypes[k] == "int":
			PValues = append(PValues, intify(v))
		}
	}
	obj["PropertyValues"] = PValues
	obj["Output"] = floatify(obj["Output"])
	return obj
}
func (e *Engine_t) AddObject(obj Object_t) {
	e.mu.Lock()
	var id int
	id = intify(obj["Id"])
	obj["Id"] = id
	obj["process"] = processors[stringify(obj["Type"])]
	obj = sanitize(obj)
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

func (e *Engine_t) EditObject(new_states StateChange) {
	id := new_states.Id
	e.mu.Lock()
	obj, ok := e.Objects[id]
	if ok {
		for k, v := range new_states.State {
			(*obj)[k] = v
		}
		var newobj Object_t
		newobj = sanitize(*obj)
		e.Objects[id] = &newobj
		e.Save()
	} else {
		fmt.Println("Edit: Unknown object")
	}
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
			var s StateChange
			switch v := event.Data.(type) {
			case map[string]interface{}:
				s = StateChange{Id: int(v["Id"].(float64)), State: v["State"].(map[string]interface{})}
			case StateChange:
				s = v
			}
			e.EditObject(s)
			fmt.Println("edit recv")
		case event.Type == "del":
			var id int
			data := event.Data.(map[string]interface{})
			id = intify(data["Id"])
			e.DeleteObject(id)
			fmt.Println("del recv")
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
	//go testPublish()
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
