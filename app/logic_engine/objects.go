package logic_engine

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
	//"timecl/app/network_manager"
	"timecl/app/logger"
)

var processors = make(map[string]processor)

func init() {
	processors["guide"] = ProcessGuide
	processors["binput"] = ProcessBinput
	processors["ainput"] = ProcessAinput
	processors["boutput"] = ProcessBoutput
	processors["aoutput"] = ProcessAoutput
	processors["notgate"] = ProcessNotGate
	processors["andgate"] = ProcessAndGate
	processors["orgate"] = ProcessOrGate
	processors["xorgate"] = ProcessXorGate
	processors["mult"] = ProcessMult
	processors["div"] = ProcessDiv
	processors["add"] = ProcessAdd
	processors["sub"] = ProcessSub
	processors["power"] = ProcessPower
	processors["sine"] = ProcessSine
	processors["cosine"] = ProcessCosine

	processors["agtb"] = ProcessAGTB
	processors["agteb"] = ProcessAGTEB
	processors["altb"] = ProcessALTB
	processors["alteb"] = ProcessALTEB
	processors["aeqb"] = ProcessAEQB
	processors["aneqb"] = ProcessANEQB

	processors["xyscope"] = ProcessXYscope
	//processors["block"] = 
	//processors["vbar"] = 
	//processors["hbar"] = 
	processors["timebase"] = ProcessTimeBase
	processors["timerange"] = ProcessTimeRange
	processors["timer"] = ProcessTimer
	processors["delay"] = ProcessDelay
	processors["conversion"] = ProcessConversion
	processors["logger"] = ProcessLogger
	processors["alert"] = ProcessAlert

	go func() {
		for {
			<-time.After(1000 * time.Millisecond)
			tbmu.Lock()
			tick += 1
			tbmu.Unlock()
		}
	}()

}

type processor func(o Object_t, objs map[int]Object_t, iteration int) error

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
	fmt.Println(output)
}

/*func (o *Object_t) Process(Objects map[int]*Object_t) error {
	return nil
}*/

func (o Object_t) AssignOutput(objs map[int]Object_t, terminal int) error {
	terms, ok := o["Terminals"].([]interface{})
	if !ok {
		return errors.New("No terminal list/terminal list of improper type.")
	}
	terminal64, ok := terms[terminal].(float64)
	if !ok {
		return errors.New("Specified terminal does not exist.")
	}
	obj, ok := objs[int(terminal64)]
	if !ok {
		return errors.New("The specified object does not exist.")
	}
	output64, ok := o["Output"].(float64)
	if !ok {
		return errors.New("No output value, or value is of improper type.")
	}
	obj["NextOutput"] = output64
	return nil
}

func (o Object_t) CheckTerminals(count int) error {
	iterms, ok := o["Terminals"]
	if !ok {
		return errors.New("No terminal list.")
	}
	terms, ok := iterms.([]interface{})
	if !ok {
		return errors.New("Terminal list of unknown type.")
	}
	if len(terms) < count {
		return fmt.Errorf("Invalid Terminals for obj type: %v, Id: %v", o["Type"], o["Id"])
	}
	return nil
}

func (o Object_t) GetTerminal(Objects map[int]Object_t, term int) (float64, error) {
	terms, ok := o["Terminals"].([]interface{})
	if !ok {
		return 0, errors.New("Terminals list of unknown type.")
	}
	terminal64, ok := terms[term].(float64)
	if !ok {
		return 0, errors.New("Specified Terminal does not exist or is of improper type.")
	}
	theterm := int(terminal64)
	obj, ok := Objects[theterm]
	if !ok {
		return 0, errors.New("Specified object does not exist.")
	}
	output64, ok := obj["Output"].(float64)
	if !ok {
		return 0, errors.New("No Output value or Output is of improper type.")
	}
	return output64, nil
}

func (o Object_t) GetProperty(name string) (interface{}, error) {
	PCount := o["PropertyCount"].(int)
	if PCount <= 0 {
		return nil, fmt.Errorf("Property %s not found.", name)
	}
	names := o["PropertyNames"].([]interface{})
	for ii := 0; ii < PCount; ii++ {
		if stringify(names[ii]) == name {
			valList, ok := o["PropertyValues"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("Property value list is of improper type.")
			}
			if ii >= len(valList) {
				return nil, fmt.Errorf("Specified property value is not in list.")
			}
			return valList[ii], nil
		}
	}
	return nil, fmt.Errorf("Property %s not found.", name)
}

func ProcessGuide(o Object_t, Objects map[int]Object_t, iteration int) error {
	source := int(o["Source"].(int))
	if source < 0 {
		return nil
	}
	o["NextOutput"] = Objects[source]["Output"]
	return nil
}

func ProcessBinput(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	port_value, ok := o["PortValue"].(float64)
	if ok {
		o["Output"] = port_value
	} else {
		o["Output"] = float64(-99)
	}

	o["NextOutput"] = o["Output"]
	err = o.AssignOutput(Objects, 0)
	return err
}

func ProcessAinput(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = o["Output"]
	port_value, ok := o["PortValue"].(float64)
	if ok {
		o["NextOutput"] = float64(port_value*(1.0/(65536.0/math.Abs(min-max))) + min)
	}

	err = o.AssignOutput(Objects, 0)
	return err
}

func ProcessBoutput(o Object_t, Objects map[int]Object_t, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o["NextOutput"] = value
	return nil
}

func ProcessAoutput(o Object_t, Objects map[int]Object_t, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o["NextOutput"] = value
	return nil
}

func ProcessNotGate(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	if input > 0 {
		o["NextOutput"] = float64(0)
	} else {
		o["NextOutput"] = float64(1)
	}
	err = o.AssignOutput(Objects, 1)
	return err
}

func ProcessAndGate(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessOrGate(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func xor(cond1, cond2 bool) bool {
	return (cond1 || cond2) && !(cond1 && cond2)
}

func ProcessXorGate(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessMult(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = in_a * in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessDiv(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = in_a / in_b
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessAdd(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = in_a + in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessSub(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = in_a - in_b
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessPower(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = math.Pow(in_a, in_b)
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessSine(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o["NextOutput"] = math.Sin(in_a)
	err = o.AssignOutput(Objects, 1)
	return err
}

func ProcessCosine(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if err = o.CheckTerminals(2); err != nil {
		return err
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	o["NextOutput"] = math.Cos(in_a)
	err = o.AssignOutput(Objects, 1)
	return err
}

func ProcessAGTB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessAGTEB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessALTB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessALTEB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessAEQB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

func ProcessANEQB(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 2)
	return err
}

var tbmu sync.Mutex
var tick float64

func ProcessTimeBase(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	tbmu.Lock()
	o["NextOutput"] = tick //float64(time.Now().Unix())
	tbmu.Unlock()
	err = o.AssignOutput(Objects, 0)
	return err
}

func ProcessXYscope(o Object_t, Objects map[int]Object_t, iteration int) error {
	if err := o.CheckTerminals(2); err != nil {
		return err
	}
	return nil
}

func ProcessTimeRange(o Object_t, Objects map[int]Object_t, iteration int) error {
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
		o["NextOutput"] = float64(0)
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
		o["NextOutput"] = float64(0)
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
		o["NextOutput"] = float64(0)
		err2 := o.AssignOutput(Objects, 0)
		if err2 != nil {
			return fmt.Errorf("Two errors encountered:", err, err2)
		} else {
			return err
		}
	}
	off_time = off_time.AddDate(year, m-1, day-1)

	if current_time.After(on_time) && current_time.Before(off_time) {
		o["NextOutput"] = float64(1)
	} else {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 0)
	return err
}

func ProcessTimer(o Object_t, Objects map[int]Object_t, iteration int) error {
	var err error
	if iteration != 0 {
		return nil
	}
	if err = o.CheckTerminals(1); err != nil {
		return err
	}
	var start int64
	var ok bool
	_, ok = o["_timer_start"]
	if ok {
		start = o["_timer_start"].(int64)
	}
	if !ok {
		start = time.Now().UTC().Unix()
		o["_timer_start"] = start
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
		o["NextOutput"] = float64(1)
	} else if modsecs >= on_secs {
		o["NextOutput"] = float64(0)
	}
	err = o.AssignOutput(Objects, 0)
	return err
}

func ProcessDelay(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	current_delay_start, ok := o["_current_delay"].(time.Time)
	if !ok {
		current_delay_start = current_time
		if input > 0 && delay > 0 {
			o["_current_delay"] = current_delay_start
			return nil
		}
	}
	delay_end := current_delay_start.Add(time.Duration(delay) * time.Second)

	if current_time.Equal(delay_end) || current_time.After(delay_end) {
		current_min_start, ok := o["_current_min_on"].(time.Time)
		if !ok {
			current_min_start = current_time
			if min > 0 {
				o["_current_min_on"] = current_min_start
			}
		}
		min_end := current_min_start.Add(time.Duration(min) * time.Second)

		if current_time.Equal(min_end) || current_time.After(min_end) {
			if input > 0 {
				o["NextOutput"] = float64(1)
			} else {
				o["NextOutput"] = float64(0)
				o["_current_delay"] = nil
				o["_current_min_on"] = nil
			}
		} else {
			o["NextOutput"] = float64(1)
		}
	} else {
		if !(input > 0) {
			o["_current_delay"] = nil
			o["_current_min_on"] = nil
		}
		o["NextOutput"] = float64(0)
	}

	err = o.AssignOutput(Objects, 1)
	return err
}

func ProcessConversion(o Object_t, Objects map[int]Object_t, iteration int) error {
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
	o["NextOutput"] = a*(input*input) + b*input + c
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

func ProcessLogger(o Object_t, Objects map[int]Object_t, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}

	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	objid, ok := o["Id"].(int)
	if !ok {
		return errors.New("Object Id doesn't exist or is of improper type.")
	}
	min, ok := o["_min_value"].(float64)
	if !ok {
		min = input
		o["_min_value"] = min
	}
	max, ok := o["_max_value"].(float64)
	if !ok {
		max = input
		o["_max_value"] = max
	}
	avg, ok := o["_avg_data"].([]float64)
	if !ok {
		avg = []float64{}
		o["_avg_data"] = avg
	}
	ifreq, _ := o.GetProperty("frequency")
	freq, ok := ifreq.(float64)
	if !ok {
		freq = 300
	}

	current := time.Now()

	next_timeslot, ok := o["_next_timeslot"].(time.Time)
	if !ok {
		_, next_timeslot = getSurroundingTimeslots(current, freq)
		o["_next_timeslot"] = next_timeslot
	}
	if current.After(next_timeslot) {
		var calc_avg float64
		for _, v := range avg {
			calc_avg += v
		}
		calc_avg = calc_avg / float64(len(avg))

		prev_timeslot, next_timeslot := getSurroundingTimeslots(current, freq)
		lEvent := logger.LoggingData{Time: prev_timeslot,
			ObjectId: objid,
			Min:      min,
			Max:      max,
			Avg:      calc_avg}
		logger.Publish(logger.Event{"capture", lEvent})

		avg = []float64{}
		min = input
		max = input
		o["_avg_data"] = avg
		o["_min_value"] = min
		o["_max_value"] = max
		o["_next_timeslot"] = next_timeslot
	}
	if input < min {
		min = input
		o["_min_value"] = min
	}
	if input > max {
		max = input
		o["_max_value"] = max
	}
	if iteration == 0 {
		avg = append(avg, input)
		o["_avg_data"] = avg
	}
	o["NextOutput"] = input
	return nil
}

func ProcessAlert(o Object_t, Objects map[int]Object_t, iteration int) error {
	if err := o.CheckTerminals(1); err != nil {
		return err
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		return err
	}
	current_time := time.Now()
	alert_event_start, ok := o["_alert_event_start"].(time.Time)
	var start = false
	var stop = false
	if !ok {
		alert_event_start = current_time
		if input > 0 {
			start = true
			o["_alert_event_start"] = alert_event_start
		}
	} else {
		if input <= 0 {
			stop = true
			delete(o, "_alert_event_start")
		}
	}
	if start || stop {
		objid, ok := o["Id"].(int)
		if !ok {
			return errors.New("Object Id doesn't exist or is of improper type.")
		}
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
					ObjectId:   objid,
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
					ObjectId:   objid,
					Subject:    subject,
					EventText:  eventText,
					Recipients: recipients}
				logger.Publish(logger.Event{"alert", aEvent})
			}
		}

	}
	o["NextOutput"] = input
	return nil
}
