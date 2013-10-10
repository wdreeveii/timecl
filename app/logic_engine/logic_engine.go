package logic_engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/robfig/revel"
	"io/ioutil"
	"sync"
	"time"
)

type processor func(o *Object_t, objs map[int]*Object_t)

type Object_t struct {
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
}

func (o Object_t) Display() {
	fmt.Printf("ID %4d  ", o.Id)
	fmt.Printf("Type %10s  ", o.Type)
	fmt.Printf("Source %3d  ", o.Source)
	fmt.Printf("Output %10f  ", o.Output)
	fmt.Printf("Terminals: ")
	for _, val := range o.Terminals {
		fmt.Printf("%d ", val)
	}
	fmt.Printf("\n")
}

func (o Object_t) Process(objects map[int]*Object_t) {
}

func (o Object_t) AssignOutput(objs map[int]*Object_t, terminal int) {
	objs[o.Terminals[terminal]].NextOutput = o.Output
}

func (o Object_t) CheckTerminals(count int) bool {
	if len(o.Terminals) < count {
		fmt.Println("Invalid ", o.Type)
		return true
	}
	return false
}
func (o Object_t) GetTerminal(objects map[int]*Object_t, term int) float64 {
	return objects[o.Terminals[term]].Output
}

func (o Object_t) GetProperty(name string) string {
	if o.PropertyCount <= 0 {
		return ""
	}
	for ii := 0; ii < o.PropertyCount; ii++ {
		if o.PropertyNames[ii] == name {
			return o.PropertyValues[ii]
		}
	}
	fmt.Println("Unable to find property ", name, " for ", o.Type)
	return ""
}

type Engine_t struct {
	mu               sync.Mutex
	objects          map[int]*Object_t
	index            int
	update_rate      float32
	solve_iterations int
}

func (e *Engine_t) Init() {
	e.update_rate = 10
	e.solve_iterations = 10
	e.objects = make(map[int]*Object_t)
	go e.Start()
}

func (e *Engine_t) Start() {
	e.LoadObjects()
	e.printObjects()
	e.Run()
}

func (e *Engine_t) Run() {
	for {
		e.mu.Lock()
		for ii := 0; ii < e.solve_iterations; ii++ {
			e.Process()
		}
		e.mu.Unlock()
		e.SetOutputs()
		e.printObjects()
		time.Sleep(time.Duration(1000/e.update_rate) * time.Millisecond)
	}
}

func (e *Engine_t) Process() {
	for _, val := range e.objects {
		val.process(val, e.objects)
	}

	for _, val := range e.objects {
		val.Output = val.NextOutput
	}
}

func (e *Engine_t) Save() {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	enc.Encode(e)
	err := ioutil.WriteFile("gob", m.Bytes(), 0600)
	if err != nil {
		panic(err)
	}
}

func (e *Engine_t) LoadObjects() {
	/*n, err := ioutil.ReadFile("gob")
	if err != nil {
		panic(err)
	}
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)
	err = dec.Decode(e)
	if err != nil {
		panic(err)
	}*/
}

func (e *Engine_t) printObjects() {
	e.mu.Lock()
	for _, val := range e.objects {
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
	for _, val := range e.objects {
		states = append(states, State_t{Id: val.Id, Output: val.Output})
	}
	e.mu.Unlock()
	return states
}

func (e *Engine_t) HookObject(id int, source int) {
	e.mu.Lock()
	e.objects[id].Source = source
	e.mu.Unlock()
}

func (e *Engine_t) UnhookObject(id int) {
	e.mu.Lock()
	e.objects[id].Source = -1
	e.mu.Unlock()
}

func (e *Engine_t) ListObjects() []Object_t {
	e.mu.Lock()
	objs := make([]Object_t, 0, len(e.objects))
	for _, val := range e.objects {
		objs = append(objs, *val)
	}
	e.mu.Unlock()
	return objs
}

func (e *Engine_t) AddObject(objtype string,
	x_pos int,
	y_pos int,
	x_size int,
	y_size int,
	attached int,
	dir int,
	property_count int,
	property_names []string,
	property_types []string,
	property_values []string) int {

	var obj = Object_t{Type: objtype, Source: -1,
		Xpos: x_pos, Ypos: y_pos,
		Xsize: x_size, Ysize: y_size,
		PropertyCount:  property_count,
		PropertyNames:  property_names,
		PropertyValues: property_values,
		process:        processors[objtype]}
	fmt.Println(property_names)
	fmt.Println(property_values)
	e.mu.Lock()
	obj_index := e.index
	obj.Id = e.index
	e.objects[e.index] = &obj
	e.index += 1
	e.mu.Unlock()
	fmt.Println("newid: ", obj_index, " ", objtype)
	return obj_index
}

func (e *Engine_t) DeleteObject(id int) {
	e.mu.Lock()
	delete(e.objects, id)
	e.mu.Unlock()
}

func (e *Engine_t) MoveObject(id, x_pos, y_pos int) {
	fmt.Println("Move: ", id, x_pos, y_pos)
	e.mu.Lock()
	e.objects[id].Xpos = x_pos
	e.objects[id].Ypos = y_pos
	e.mu.Unlock()
}

func (e *Engine_t) SetGuides(id int, guide int) {
	e.mu.Lock()
	e.objects[id].Terminals = append(e.objects[id].Terminals, guide)
	e.mu.Unlock()
}

func (e *Engine_t) SetOutput(id int, output float64) {
	e.mu.Lock()
	fmt.Println("Setting output...", output)
	e.objects[id].Output = output
	e.mu.Unlock()
}
func (e *Engine_t) SetProperties(id int, property_count int,
	property_names []string, property_types []string, property_values []string) {
	e.mu.Lock()
	e.objects[id].PropertyNames = property_names
	e.objects[id].PropertyTypes = property_types
	e.objects[id].PropertyValues = property_values
	e.mu.Unlock()
}

func Init() {
	fmt.Println("engine start")
}

func init() {
	revel.OnAppStart(Init)
}
