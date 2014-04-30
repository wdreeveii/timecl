package logic_engine

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
	"timecl/app/logger"
	"timecl/app/network_manager"
)

var processors = make(map[Type_t]processor)
var inits = make(map[Type_t]initfunc)

func init() {
	processors[Type_t("guide")] = ProcessGuide
	processors[Type_t("binput")] = ProcessBinput
	processors[Type_t("ainput")] = ProcessAinput
	processors[Type_t("boutput")] = ProcessBoutput
	processors[Type_t("aoutput")] = ProcessAoutput
	processors[Type_t("notgate")] = ProcessNotGate
	processors[Type_t("andgate")] = ProcessAndGate
	processors[Type_t("orgate")] = ProcessOrGate
	processors[Type_t("xorgate")] = ProcessXorGate
	processors[Type_t("mult")] = ProcessMult
	processors[Type_t("div")] = ProcessDiv
	processors[Type_t("add")] = ProcessAdd
	processors[Type_t("sub")] = ProcessSub
	processors[Type_t("power")] = ProcessPower
	processors[Type_t("sine")] = ProcessSine
	processors[Type_t("cosine")] = ProcessCosine

	processors[Type_t("agtb")] = ProcessAGTB
	processors[Type_t("agteb")] = ProcessAGTEB
	processors[Type_t("altb")] = ProcessALTB
	processors[Type_t("alteb")] = ProcessALTEB
	processors[Type_t("aeqb")] = ProcessAEQB
	processors[Type_t("aneqb")] = ProcessANEQB

	processors[Type_t("xyscope")] = ProcessXYscope

	processors[Type_t("timebase")] = ProcessTimeBase
	processors[Type_t("timerange")] = ProcessTimeRange
	processors[Type_t("timer")] = ProcessTimer
	processors[Type_t("delay")] = ProcessDelay
	processors[Type_t("conversion")] = ProcessConversion
	processors[Type_t("logger")] = ProcessLogger
	processors[Type_t("alert")] = ProcessAlert

	inits[Type_t("guide")] = InitGuide
	inits[Type_t("binput")] = InitBinput
	inits[Type_t("ainput")] = InitAinput
	inits[Type_t("boutput")] = InitBoutput
	inits[Type_t("aoutput")] = InitAoutput
	inits[Type_t("notgate")] = InitNotGate
	inits[Type_t("andgate")] = InitAndGate
	inits[Type_t("orgate")] = InitOrGate
	inits[Type_t("xorgate")] = InitXorGate
	inits[Type_t("mult")] = InitMult
	inits[Type_t("div")] = InitDiv
	inits[Type_t("add")] = InitAdd
	inits[Type_t("sub")] = InitSub
	inits[Type_t("power")] = InitPower
	inits[Type_t("sine")] = InitSine
	inits[Type_t("cosine")] = InitCosine

	inits[Type_t("agtb")] = InitAGTB
	inits[Type_t("agteb")] = InitAGTEB
	inits[Type_t("altb")] = InitALTB
	inits[Type_t("alteb")] = InitALTEB
	inits[Type_t("aeqb")] = InitAEQB
	inits[Type_t("aneqb")] = InitANEQB

	inits[Type_t("xyscope")] = InitXYscope

	inits[Type_t("timebase")] = InitTimeBase
	inits[Type_t("timerange")] = InitTimeRange
	inits[Type_t("timer")] = InitTimer
	inits[Type_t("delay")] = InitDelay
	inits[Type_t("conversion")] = InitConversion
	inits[Type_t("logger")] = InitLogger
	inits[Type_t("alert")] = InitAlert

	go func() {
		for {
			<-time.After(1000 * time.Millisecond)
			tbmu.Lock()
			tick += 1
			tbmu.Unlock()
		}
	}()

}

type processor func(o *Object_t, objs ObjectList, iteration int) error
type initfunc func() Object_t

type Id_t int
type Type_t string
type Dim_t int
type Value_t network_manager.Value_t
type SignalGood bool
type Terminals_t []Id_t
type PropertyCount_t int
type PropertyNames_t []interface{}
type PropertyValues_t []interface{}
type PropertyTypes_t []interface{}
type StateData_t map[string]interface{}
type Attached_t int
type Dir_t int

//type Object_t map[string]interface{}

type Object_t struct {
	Id              Id_t
	Type            Type_t
	Xpos            Dim_t
	Ypos            Dim_t
	Xsize           Dim_t
	Ysize           Dim_t
	Output          Value_t
	NextOutput      Value_t
	SignalGood      SignalGood
	PortValue       Value_t
	Terminals       Terminals_t
	Source          Id_t
	PropertyCount   PropertyCount_t
	PropertyNames   PropertyNames_t
	PropertyValues  PropertyValues_t
	PropertyTypes   PropertyTypes_t
	StateData       StateData_t
	Attached        Attached_t
	Dir             Dir_t
	process         processor
	ShowOutput      bool
	ShowAnalog      bool
	ShowName        string
	inputTermCount  int
	outputTermCount int
}

func (o Object_t) Copy() Object_t {
	var n Object_t = o
	n.PropertyNames = make(PropertyNames_t, len(n.PropertyNames), len(n.PropertyNames))
	n.PropertyValues = make(PropertyValues_t, len(n.PropertyValues), len(n.PropertyValues))
	n.PropertyTypes = make(PropertyTypes_t, len(n.PropertyTypes), len(n.PropertyTypes))

	copy(n.PropertyNames, o.PropertyNames)
	copy(n.PropertyValues, o.PropertyValues)
	copy(n.PropertyTypes, o.PropertyTypes)

	n.StateData = nil
	return n
}

const (
	DirNone  = Dir_t(0)
	DirUp    = Dir_t(1)
	DirDown  = Dir_t(2)
	DirLeft  = Dir_t(3)
	DirRight = Dir_t(4)
)

func makeObject_t() Object_t {
	var obj Object_t
	obj.Xsize = 50
	obj.Ysize = 50

	obj.Type = Type_t("None")

	obj.Dir = DirNone
	obj.Attached = -1
	obj.Source = -1

	obj.Terminals = make(Terminals_t, 0)
	obj.PropertyNames = make(PropertyNames_t, 0)
	obj.PropertyValues = make(PropertyValues_t, 0)
	obj.PropertyTypes = make(PropertyTypes_t, 0)
	obj.StateData = make(StateData_t)
	return obj
}

type ObjectList map[Id_t]*Object_t

func (o Object_t) Display() {
	var output string
	output += fmt.Sprintf("ID %4d  ", o.Id)
	output += fmt.Sprintf("Type %10s  ", o.Type)
	output += fmt.Sprintf("Source %3d  ", o.Source)
	output += fmt.Sprintf("Output %10f  ", o.Output)
	output += fmt.Sprintf("Terminals: ")
	for _, val := range o.Terminals {
		output += fmt.Sprintf("%d ", val)
	}
	fmt.Println(output)
}

/*func (o *Object_t) Process(Objects map[int]*Object_t) error {
	return nil
}*/

func (o Object_t) AssignOutput(objs ObjectList, terminal int) error {
	if terminal < len(o.Terminals) {
		term := o.Terminals[terminal]
		obj, exists := objs[term]
		if exists {
			obj.NextOutput = o.NextOutput
			return nil
		} else {
			return errors.New("The specified object does not exist.")
		}
	} else {
		return errors.New("Specified terminal does not exist.")
	}
}

func (o Object_t) CheckTerminals(count int) error {
	if len(o.Terminals) < count {
		return fmt.Errorf("Invalid Terminals for obj type: %v, Id: %v", o.Type, o.Id)
	}
	return nil
}

func (o Object_t) GetTerminal(Objects ObjectList, term int) (Value_t, error) {
	if term < len(o.Terminals) {
		terminal_id := o.Terminals[term]
		obj, exists := Objects[terminal_id]
		if exists {
			return obj.Output, nil
		} else {
			return 0, errors.New("Specified object does not exist.")
		}
	} else {
		return 0, errors.New("Specified terminal does not exist.")
	}
}

func (o *Object_t) addProperty(name string, prop_type string, default_value interface{}) {
	o.PropertyNames = append(o.PropertyNames, name)
	o.PropertyTypes = append(o.PropertyTypes, prop_type)
	o.PropertyValues = append(o.PropertyValues, default_value)
	o.PropertyCount++
}

func (o Object_t) GetProperty(name string) (interface{}, error) {
	for ii := PropertyCount_t(0); ii < o.PropertyCount; ii++ {
		if o.PropertyNames[ii] == name {
			if int(ii) < len(o.PropertyValues) {
				val := o.PropertyValues[ii]
				return val, nil
			} else {
				return nil, fmt.Errorf("Specified property value is not in list.")
			}
		}
	}
	return nil, fmt.Errorf("Property %s not found.", name)
}

func InitGuide() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 10
	obj.Ysize = 10
	obj.process = ProcessGuide
	return obj
}

func ProcessGuide(o *Object_t, Objects ObjectList, iteration int) error {
	Source, exists := Objects[o.Source]
	if exists {
		o.NextOutput = Source.Output
	}
	return nil
}

func InitBinput() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 60
	obj.Ysize = 60
	obj.ShowOutput = true
	obj.inputTermCount = 0
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("value", "float", 0)
	obj.addProperty("port", "port", "None")
	obj.process = ProcessBinput
	return obj
}

func ProcessBinput(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	if o.SignalGood {
		o.Output = o.PortValue
	}
	o.NextOutput = o.Output
	err = o.AssignOutput(Objects, 0)
	return err
}

func InitAinput() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 60
	obj.Ysize = 60
	obj.ShowOutput = true
	obj.inputTermCount = 0
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("value", "float", 0)
	obj.addProperty("port", "port", "None")
	obj.addProperty("Auto scale - Max", "float", 5)
	obj.addProperty("Auto scale - Min", "float", 0)
	obj.process = ProcessAinput
	return obj
}

func ProcessAinput(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	imin, err := o.GetProperty("Auto scale - Min")
	if err != nil {
		logger.PublishOneError(err)
	}
	min, ok := imin.(float64)
	if !ok {
		min = float64(0)
	}
	imax, err := o.GetProperty("Auto scale - Max")
	if err != nil {
		logger.PublishOneError(err)
	}
	max, ok := imax.(float64)
	if !ok {
		max = float64(5)
	}
	o.NextOutput = o.Output
	if o.SignalGood {
		delta := Value_t(math.Abs(min - max))
		base_delta := Value_t(65536.0) / delta
		ratio := Value_t(1.0) / base_delta
		o.NextOutput = o.PortValue*ratio + Value_t(min)
	}

	err = o.AssignOutput(Objects, 0)
	return err
}

func InitBoutput() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 80
	obj.Ysize = 30
	obj.ShowOutput = true
	obj.inputTermCount = 1
	obj.outputTermCount = 0
	obj.addProperty("name", "string", "")
	obj.addProperty("port", "port", "None")
	obj.process = ProcessBoutput
	return obj
}

func ProcessBoutput(o *Object_t, Objects ObjectList, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o.NextOutput = value
	return nil
}

func InitAoutput() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 80
	obj.Ysize = 30
	obj.ShowAnalog = true
	obj.inputTermCount = 1
	obj.outputTermCount = 0
	obj.addProperty("name", "string", "")
	obj.addProperty("port", "port", "None")
	obj.process = ProcessAoutput
	return obj
}

func ProcessAoutput(o *Object_t, Objects ObjectList, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o.NextOutput = value
	return nil
}

func InitNotGate() Object_t {
	var obj = makeObject_t()
	obj.ShowOutput = true
	obj.inputTermCount = 1
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessNotGate
	return obj
}

func ProcessNotGate(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	if input > 0 {
		o.NextOutput = Value_t(0)
	} else {
		o.NextOutput = Value_t(1)
	}
	err = o.AssignOutput(Objects, 1)
	return err
}

func InitAndGate() Object_t {
	var obj = makeObject_t()
	obj.ShowOutput = true
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessAndGate
	return obj
}

func ProcessAndGate(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}

	if in_a > 0 && in_b > 0 {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitOrGate() Object_t {
	var obj = makeObject_t()
	obj.ShowOutput = true
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessOrGate
	return obj
}

func ProcessOrGate(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a > 0 || in_b > 0 {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitXorGate() Object_t {
	var obj = makeObject_t()
	obj.ShowOutput = true
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessXorGate
	return obj
}

func xor(cond1, cond2 bool) bool {
	return (cond1 || cond2) && !(cond1 && cond2)
}

func ProcessXorGate(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if xor((in_a > 0), (in_b > 0)) {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitMult() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Mult"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessMult
	return obj
}

func ProcessMult(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	o.NextOutput = in_a * in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitDiv() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Div"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessDiv
	return obj
}

func ProcessDiv(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err := o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_b != 0 {
		o.NextOutput = in_a / in_b
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitAdd() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Add"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessAdd
	return obj
}

func ProcessAdd(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	o.NextOutput = in_a + in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitSub() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Sub"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessSub
	return obj
}

func ProcessSub(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	o.NextOutput = in_a - in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitPower() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Power"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessPower
	return obj
}

func ProcessPower(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	o.NextOutput = Value_t(math.Pow(float64(in_a), float64(in_b)))
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitSine() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Sine"
	obj.inputTermCount = 1
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessSine
	return obj
}

func ProcessSine(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o.NextOutput = Value_t(math.Sin(float64(in_a)))
	err = o.AssignOutput(Objects, 1)
	return err
}

func InitCosine() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Cosine"
	obj.inputTermCount = 1
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessCosine
	return obj
}

func ProcessCosine(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o.NextOutput = Value_t(math.Cos(float64(in_a)))
	err = o.AssignOutput(Objects, 1)
	return err
}

func InitAGTB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A > B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessAGTB
	return obj
}

func ProcessAGTB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a > in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitAGTEB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A >= B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessAGTEB
	return obj
}

func ProcessAGTEB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a >= in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitALTB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A < B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessALTB
	return obj
}

func ProcessALTB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a < in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitALTEB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A <= B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessALTEB
	return obj
}

func ProcessALTEB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a <= in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitAEQB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A == B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessAEQB
	return obj
}

func ProcessAEQB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a == in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitANEQB() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "A != B"
	obj.inputTermCount = 2
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessANEQB
	return obj
}

func ProcessANEQB(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err := o.CheckTerminals(3); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		return err
	}
	if in_a != in_b {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func InitTimeBase() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 60
	obj.Ysize = 60
	obj.ShowOutput = true
	obj.inputTermCount = 0
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.process = ProcessTimeBase
	tbmu.Lock()
	obj.Output = Value_t(tick)
	tbmu.Unlock()
	return obj
}

var tbmu sync.Mutex
var tick float64

func ProcessTimeBase(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	tbmu.Lock()
	o.NextOutput = Value_t(tick) //float64(time.Now().Unix())
	tbmu.Unlock()
	err = o.AssignOutput(Objects, 0)
	return err
}

func InitXYscope() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 200
	obj.Ysize = 200
	obj.inputTermCount = 2
	obj.outputTermCount = 0
	obj.addProperty("name", "string", "")
	obj.process = ProcessXYscope
	return obj
}

func ProcessXYscope(o *Object_t, Objects ObjectList, iteration int) error {
	if err := o.CheckTerminals(2); err != nil {
		return err
	}
	return nil
}

func InitTimeRange() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 80
	obj.Ysize = 40
	obj.ShowOutput = true
	obj.inputTermCount = 0
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("on", "time", "8:00")
	obj.addProperty("off", "time", "18:00")
	obj.addProperty("timezone", "timezone", "")
	obj.process = ProcessTimeRange
	return obj
}

func ProcessTimeRange(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if iteration != 0 {
		return nil
	}
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	ion, err := o.GetProperty("on")
	if err != nil {
		return err
	}
	on, ok := ion.(string)
	if !ok {
		return errors.New("on time property is of improper type.")
	}
	ioff, err := o.GetProperty("off")
	if err != nil {
		return err
	}
	off, ok := ioff.(string)
	if !ok {
		return errors.New("off time property is of improper type.")
	}
	itimezone, err := o.GetProperty("timezone")
	if err != nil {
		return err
	}
	timezone, ok := itimezone.(string)
	if !ok {
		return errors.New("timezone property is of improper type.")
	}
	if timezone == "" {
		timezone = "Local"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		o.NextOutput = Value_t(0)
		err2 := o.AssignOutput(Objects, 0)
		if err2 != nil {
			return fmt.Errorf("Two errors encountered:", err, err2)
		} else {
			return err
		}
	}
	var current_time = time.Now()
	var year, month, day = current_time.Date()
	m := int(month)
	on_time, err := time.ParseInLocation("15:04", on, loc)
	if err != nil {
		o.NextOutput = Value_t(0)
		err2 := o.AssignOutput(Objects, 0)
		if err2 != nil {
			fmt.Errorf("Two errors encountered:", err, err2)
		} else {
			return err
		}
	}
	on_time = on_time.AddDate(year, m-1, day-1)
	off_time, err := time.ParseInLocation("15:04", off, loc)
	if err != nil {
		o.NextOutput = Value_t(0)
		err2 := o.AssignOutput(Objects, 0)
		if err2 != nil {
			return fmt.Errorf("Two errors encountered:", err, err2)
		} else {
			return err
		}
	}
	off_time = off_time.AddDate(year, m-1, day-1)

	if current_time.After(on_time) && current_time.Before(off_time) {
		o.NextOutput = Value_t(1)
	} else {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 0)
	return err
}

func InitTimer() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 80
	obj.Ysize = 60
	obj.ShowOutput = true
	obj.inputTermCount = 0
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("on duration", "time", "2s")
	obj.addProperty("off duration", "time", "2s")
	obj.process = ProcessTimer
	return obj
}

func ProcessTimer(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if iteration != 0 {
		return nil
	}
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	var start int64
	var ok bool
	_, ok = o.StateData["_timer_start"]
	if ok {
		start = o.StateData["_timer_start"].(int64)
	} else {
		start = time.Now().UTC().Unix()
		o.StateData["_timer_start"] = start
	}
	now := time.Now().UTC().Unix()
	ion, err := o.GetProperty("on duration")
	if err != nil {
		return err
	}
	on, ok := ion.(string)
	if !ok {
		return errors.New("on duration property is of improper type.")
	}
	ioff, err := o.GetProperty("off duration")
	if err != nil {
		return err
	}
	off, ok := ioff.(string)
	if !ok {
		return errors.New("off duration property is of improper type.")
	}
	on_dur, err := time.ParseDuration(on)
	if err != nil {
		return err
	}
	off_dur, err := time.ParseDuration(off)
	if err != nil {
		return err
	}
	on_secs := int64(on_dur / time.Second)
	off_secs := int64(off_dur / time.Second)
	//fmt.Println("on_secs", on_secs)
	//fmt.Println("off_secs", off_secs)
	modsecs := (now - start) % (on_secs + off_secs)
	if modsecs >= 0 && modsecs < on_secs {
		o.NextOutput = Value_t(1)
	} else if modsecs >= on_secs {
		o.NextOutput = Value_t(0)
	}
	err = o.AssignOutput(Objects, 0)
	return err
}

func InitDelay() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Delay"
	obj.inputTermCount = 1
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("delay", "float", 0)
	obj.addProperty("min on", "float", 0)
	obj.process = ProcessDelay
	return obj
}

func ProcessDelay(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	idelay, err := o.GetProperty("delay")
	if err != nil {
		logger.PublishOneError(err)
	}
	delay, ok := idelay.(float64)
	if !ok {
		delay = float64(0)
	}
	imin, err := o.GetProperty("min on")
	if err != nil {
		logger.PublishOneError(err)
	}
	min, ok := imin.(float64)
	if !ok {
		min = float64(0)
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	current_time := time.Now()
	current_delay_start, ok := o.StateData["_current_delay"].(time.Time)
	if !ok {
		current_delay_start = current_time
		if input > 0 && delay > 0 {
			o.StateData["_current_delay"] = current_delay_start
			return nil
		}
	}
	delay_end := current_delay_start.Add(time.Duration(delay) * time.Second)

	if current_time.Equal(delay_end) || current_time.After(delay_end) {
		current_min_start, ok := o.StateData["_current_min_on"].(time.Time)
		if !ok {
			current_min_start = current_time
			if min > 0 {
				o.StateData["_current_min_on"] = current_min_start
			}
		}
		min_end := current_min_start.Add(time.Duration(min) * time.Second)

		if current_time.Equal(min_end) || current_time.After(min_end) {
			if input > 0 {
				o.NextOutput = Value_t(1)
			} else {
				o.NextOutput = Value_t(0)
				o.StateData["_current_delay"] = nil
				o.StateData["_current_min_on"] = nil
			}
		} else {
			o.NextOutput = Value_t(1)
		}
	} else {
		if !(input > 0) {
			o.StateData["_current_delay"] = nil
			o.StateData["_current_min_on"] = nil
		}
		o.NextOutput = Value_t(0)
	}

	err = o.AssignOutput(Objects, 1)
	return err
}

func InitConversion() Object_t {
	var obj = makeObject_t()
	obj.Xsize = 140
	obj.Ysize = 60
	obj.ShowAnalog = true
	obj.ShowName = "Conversion"
	obj.inputTermCount = 1
	obj.outputTermCount = 1
	obj.addProperty("name", "string", "")
	obj.addProperty("a", "float", 0)
	obj.addProperty("b", "float", 0)
	obj.addProperty("c", "float", 0)
	obj.process = ProcessConversion
	return obj
}

func ProcessConversion(o *Object_t, Objects ObjectList, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	ia, err := o.GetProperty("a")
	if err != nil {
		logger.PublishOneError(err)
	}
	a, ok := ia.(float64)
	if !ok {
		a = float64(0)
	}
	ib, err := o.GetProperty("b")
	if err != nil {
		logger.PublishOneError(err)
	}
	b, ok := ib.(float64)
	if !ok {
		b = float64(0)
	}
	ic, err := o.GetProperty("c")
	if err != nil {
		logger.PublishOneError(err)
	}
	c, ok := ic.(float64)
	if !ok {
		c = float64(0)
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o.NextOutput = Value_t(a)*(input*input) + Value_t(b)*input + Value_t(c)
	err = o.AssignOutput(Objects, 1)
	return err
}

func getSurroundingTimeslots(current time.Time, freq float64) (time.Time, time.Time) {
	year, month, day := current.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	dfreq := time.Duration(freq) * time.Second
	current_timeslot := current.Sub(today) / dfreq
	prev_timeslot_start := today.Add((current_timeslot - 1) * dfreq)
	next_timeslot_start := today.Add((current_timeslot + 1) * dfreq)
	return prev_timeslot_start, next_timeslot_start
}

func InitLogger() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Log"
	obj.inputTermCount = 1
	obj.outputTermCount = 0
	obj.addProperty("name", "string", "")
	obj.addProperty("frequency", "float", 300)
	obj.process = ProcessLogger
	return obj
}

func ProcessLogger(o *Object_t, Objects ObjectList, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}

	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	objid := o.Id
	min, ok := o.StateData["_min_value"].(Value_t)
	if !ok {
		min = input
		o.StateData["_min_value"] = min
	}
	max, ok := o.StateData["_max_value"].(Value_t)
	if !ok {
		max = input
		o.StateData["_max_value"] = max
	}
	avg, ok := o.StateData["_avg_data"].([]Value_t)
	if !ok {
		avg = []Value_t{}
		o.StateData["_avg_data"] = avg
	}
	ifreq, _ := o.GetProperty("frequency")
	freq, ok := ifreq.(float64)
	if !ok {
		freq = 300
	}

	current := time.Now()

	next_timeslot, ok := o.StateData["_next_timeslot"].(time.Time)
	if !ok {
		_, next_timeslot = getSurroundingTimeslots(current, freq)
		o.StateData["_next_timeslot"] = next_timeslot
	}
	if current.After(next_timeslot) {
		var calc_avg Value_t
		for _, v := range avg {
			calc_avg += v
		}
		calc_avg = calc_avg / Value_t(len(avg))

		prev_timeslot, next_timeslot := getSurroundingTimeslots(current, freq)
		lEvent := logger.LoggingData{Time: prev_timeslot,
			ObjectId: int(objid),
			Min:      float64(min),
			Max:      float64(max),
			Avg:      float64(calc_avg)}
		logger.Publish(logger.Event{"capture", lEvent})

		avg = []Value_t{}
		min = input
		max = input
		o.StateData["_avg_data"] = avg
		o.StateData["_min_value"] = min
		o.StateData["_max_value"] = max
		o.StateData["_next_timeslot"] = next_timeslot
	}
	if input < min {
		min = input
		o.StateData["_min_value"] = min
	}
	if input > max {
		max = input
		o.StateData["_max_value"] = max
	}
	if iteration == 0 {
		avg = append(avg, input)
		o.StateData["_avg_data"] = avg
	}
	o.NextOutput = input
	return nil
}

func InitAlert() Object_t {
	var obj = makeObject_t()
	obj.ShowAnalog = true
	obj.ShowName = "Alert"
	obj.inputTermCount = 1
	obj.outputTermCount = 0
	obj.addProperty("name", "string", "")
	obj.addProperty("Event Text", "string", "")
	obj.addProperty("Email Recipients", "string", "")
	obj.addProperty("Notify Event Start", "string", "Yes")
	obj.addProperty("Notify Event End", "string", "Yes")
	obj.process = ProcessAlert
	return obj
}

func ProcessAlert(o *Object_t, Objects ObjectList, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	current_time := time.Now()
	alert_event_start, ok := o.StateData["_alert_event_start"].(time.Time)
	var start = false
	var stop = false
	if !ok {
		alert_event_start = current_time
		if input > 0 {
			start = true
			o.StateData["_alert_event_start"] = alert_event_start
		}
	} else {
		if input <= 0 {
			stop = true
			delete(o.StateData, "_alert_event_start")
		}
	}
	if start || stop {
		objid := o.Id
		iname, err := o.GetProperty("name")
		if err != nil {
			logger.PublishOneError(err)
		}
		name, ok := iname.(string)
		if !ok {
			name = ""
		}
		ieventText, err := o.GetProperty("Event Text")
		if err != nil {
			return err
		}
		eventText, ok := ieventText.(string)
		if !ok {
			return errors.New("EventText is of improper type.")
		}
		irecipients, err := o.GetProperty("Email Recipients")
		if err != nil {
			return err
		}
		recipients, ok := irecipients.(string)
		if !ok {
			return errors.New("Email Recipients doesn't exist or is of improper type.")
		}
		if start {
			inotify_event_start, err := o.GetProperty("Notify Event Start")
			if err != nil {
				logger.PublishOneError(err)
			}
			notify_event_start, ok := inotify_event_start.(string)
			if !ok {
				notify_event_start = "No"
			}
			if notify_event_start == "Yes" {
				var subject = "[" + name + "] "
				subject += "Triggered: " + current_time.Format(time.StampMilli)
				aEvent := logger.AlertData{Time: current_time,
					ObjectId:   int(objid),
					Subject:    subject,
					EventText:  eventText,
					Recipients: recipients}
				logger.Publish(logger.Event{"alert", aEvent})
			}
		} else if stop {
			inotify_event_end, err := o.GetProperty("Notify Event End")
			if err != nil {
				logger.PublishOneError(err)
			}
			notify_event_end, ok := inotify_event_end.(string)
			if !ok {
				notify_event_end = "No"
			}
			if notify_event_end == "Yes" {
				var subject = "[" + name + "] "
				subject += "Recovered: " + current_time.Format(time.StampMilli)
				aEvent := logger.AlertData{Time: current_time,
					ObjectId:   int(objid),
					Subject:    subject,
					EventText:  eventText,
					Recipients: recipients}
				logger.Publish(logger.Event{"alert", aEvent})
			}
		}

	}
	o.NextOutput = input
	return nil
}
