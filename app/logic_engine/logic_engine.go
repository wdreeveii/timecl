
package logic_engine

import (
	"fmt"
	"github.com/robfig/revel"
	"sync"
)

type Object_t struct {
	Id				int
	Type			string
	Xpos			int
	Ypos			int
	Xsize			int
	Ysize			int
	Output			float32
	NextOutput		float32
	TermList		string
	Terminals		[]int
	Source			int
	PropertyCount 	int
	PropertyNames 	[]string
	PropertyValues 	[]string
	PropertyTypes	[]string
	Attached		int
	Dir				int
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

func (o Object_t) Process(objects []Object_t) {
}

func (o Object_t) AssignOutput(objs []Object_t, terminal int) {
	objs[o.Terminals[terminal]].NextOutput = o.Output
}

func (o Object_t) CheckTerminals(count int) int {
	if len(o.Terminals) < count {
		fmt.Println("Invalid ", o.Type)
		return 1
	}
	return 0
}
func (o Object_t) GetTerminal(objects []Object_t, term int) float32 {
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
	mu				sync.Mutex
	Objects			[]Object_t
	UpdateRate		float32
	SolveIterations int
}

func (e Engine_t) Start () {
	e.LoadObjects()
	e.printObjects()
	e.Run()
}

func (e Engine_t) Run () {
	for {
		e.LoadObjects()
		
		for ii := 0; ii < e.SolveIterations; ii++ {
			e.Process()
		}
		e.SetOutputs()
		e.printObjects();
		//time.Sleep(
	}
}

func (e Engine_t) Process() {
	for _, val := range e.Objects {
		val.Process(e.Objects)
	}
	
	for _, val := range e.Objects {
		val.Output = val.NextOutput
	}
}

func (e Engine_t) QueryObjects() {
	
}

func (e Engine_t) LoadObjects() {
	
}

func (e Engine_t) printObjects() {
	e.mu.Lock()
	for _, val := range e.Objects {
		val.Display()
	}
	e.mu.Unlock()
	// new line?
}

func (e Engine_t) GetOutputs() {
	for { // range result list
		// if object not in list
		// return
		//e.Objects[id].Output = output_from_db
	}
}

func (e Engine_t) SetOutputs() {
	// for each object set
}

type State_t struct {
	Id	int
	Output float32
}

func (e *Engine_t) GetStates() []State_t {
	e.mu.Lock()
	var states []State_t
	for _, val := range e.Objects {
		states = append(states, State_t{Id: val.Id, Output: val.Output})
	}
	e.mu.Unlock()
	return states
}

func (e *Engine_t) ListObjects() []Object_t {
	e.mu.Lock()
	objs := make([]Object_t, len(e.Objects))
	copy(objs, e.Objects)
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
				property_names string,
				property_types string,
				property_values string) int {
	var obj = Object_t{Type: objtype, Source: -1,
						Xpos: x_pos, Ypos: y_pos,
						Xsize: x_size, Ysize: y_size,
						PropertyCount: property_count,
						PropertyNames: []string{property_names},
						PropertyValues: []string{property_values}}
	
	e.mu.Lock()
	obj_index := len(e.Objects)
	obj.Id = obj_index
	e.Objects = append(e.Objects, obj)
	e.mu.Unlock()
	return obj_index
}

func (e *Engine_t) MoveObject(id, x_pos, y_pos int) {
	fmt.Println("Move: ", id, x_pos, y_pos)
	e.mu.Lock()
	e.Objects[id].Xpos = x_pos
	e.Objects[id].Ypos = y_pos
	e.mu.Unlock()
}

func (e *Engine_t) SetGuides(id int, guide int) {
	e.mu.Lock()
	e.Objects[id].Terminals = append(e.Objects[id].Terminals, guide)
	e.mu.Unlock()
}

func engineManager() {
	//var engine Engine_t
	
	/*for {
		select {
		case obj_add_cmd, ok := <- addobj:
			if (!ok) {
				return
			}
			obj_index := len(engine.Objects)
			engine.Objects = append(engine.Objects, obj_add_cmd.obj)
			obj_add_cmd.recv <- obj_index
		case statecmd, ok := <- getstates:
			if (!ok) {
				return
			}
			states := StateResult{Num: len(engine.Objects)}
			for index, val := range engine.Objects {
				states.ObjList = append(states.ObjList, States_t{Id: index, Output: val.Output})
			}
			statecmd.recv <- states
		}
	}*/
}
func Init() {
	fmt.Println("engine start")
	go engineManager()
	/*db.Init()
	dbm := &gorp.DbMap{Db: db.Db, Dialect: gorp.SqliteDialect{}}

	init_networkconfig_table(dbm)
	
	result := GetHardwareInterfaces()
	fmt.Println("results: ", result)
	for _, config_key := range result {
		fmt.Println(config_key)
		networks, err := dbm.Select(models.NetworkConfig{}, `select * from NetworkConfig where ConfigKey = ?`, config_key)
		if err != nil {
			panic(err)
		}
		var driver driverListItem
		if len(networks) > 0 {
			driver_name := networks[0].(*models.NetworkConfig).Driver
			for index, driver_list_item := range driver_collection {
				if driver_name == driver_list_item.Name {
					driver = driver_collection[index]
				}
			}
		}
		newInterface <- interfaceItem{ConfigKey: config_key, Driver: driver}
	}*/
}

func init() {
	revel.OnAppStart(Init)
}
