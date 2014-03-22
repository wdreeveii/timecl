package logic_engine

import (
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

	go func() {
		for {
			<-time.After(1000 * time.Millisecond)
			tbmu.Lock()
			tick += 1
			tbmu.Unlock()
		}
	}()

}

func ProcessGuide(o *Object_t, Objects map[int]*Object_t, iteration int) {
	source := int((*o)["Source"].(int))
	if source < 0 {
		return
	}
	(*o)["NextOutput"] = (*Objects[source])["Output"]
}

func ProcessBinput(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	port_value, ok := (*o)["PortValue"]
	if ok {
		val, ok := port_value.(float64)
		if ok {
			(*o)["Output"] = val
		}
	}

	(*o)["NextOutput"] = (*o)["Output"]
	o.AssignOutput(Objects, 0)
}

func ProcessAinput(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	min, ok := o.GetProperty("Auto scale - Min").(float64)
	if !ok {
		min = 0
	}
	max, ok := o.GetProperty("Auto scale - Max").(float64)
	if !ok {
		max = 5
	}
	(*o)["NextOutput"] = (*o)["Output"]
	port_value, ok := (*o)["PortValue"]
	if ok {
		in, ok := port_value.(float64)
		if ok {
			(*o)["NextOutput"] = float64(in*(1.0/(65536.0/math.Abs(min-max))) + min)
		}
	}

	o.AssignOutput(Objects, 0)
}

func ProcessBoutput(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process boutput:", err)
		return
	}
	(*o)["NextOutput"] = value
}

func ProcessAoutput(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	value, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process aoutput:", err)
	}
	(*o)["NextOutput"] = value
}

func ProcessNotGate(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process not gate:", err)
		return
	}
	if input > 0 {
		(*o)["NextOutput"] = float64(0)
	} else {
		(*o)["NextOutput"] = float64(1)
	}
	o.AssignOutput(Objects, 1)
}

func ProcessAndGate(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process and gate:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process and gate:", err)
		return
	}

	if in_a > 0 && in_b > 0 {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessOrGate(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process or gate:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process or gate:", err)
		return
	}
	if in_a > 0 || in_b > 0 {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func xor(cond1, cond2 bool) bool {
	return (cond1 || cond2) && !(cond1 && cond2)
}

func ProcessXorGate(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process xor gate:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process xor gate:", err)
		return
	}
	if xor((in_a > 0), (in_b > 0)) {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessMult(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process mult:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process mult:", err)
		return
	}
	(*o)["NextOutput"] = in_a * in_b
	o.AssignOutput(Objects, 2)
}

func ProcessDiv(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process div:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process div:", err)
		return
	}
	if in_b != 0 {
		(*o)["NextOutput"] = in_a / in_b
	}
	o.AssignOutput(Objects, 2)
}

func ProcessAdd(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process add:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process add:", err)
		return
	}
	(*o)["NextOutput"] = in_a + in_b
	o.AssignOutput(Objects, 2)
}

func ProcessSub(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process subtraction:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process subtraction:", err)
		return
	}
	(*o)["NextOutput"] = in_a - in_b
	o.AssignOutput(Objects, 2)
}

func ProcessPower(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	(*o)["NextOutput"] = math.Pow(in_a, in_b)
	o.AssignOutput(Objects, 2)
}

func ProcessSine(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process sine:", err)
		return
	}
	(*o)["NextOutput"] = math.Sin(in_a)
	o.AssignOutput(Objects, 1)
}

func ProcessCosine(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process cosine:", err)
	}
	(*o)["NextOutput"] = math.Cos(in_a)
	o.AssignOutput(Objects, 1)
}

func ProcessAGTB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a > in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessAGTEB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a >= in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessALTB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a < in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessALTEB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a <= in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessAEQB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a == in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessANEQB(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(3) {
		return
	}
	in_a, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	in_b, err := o.GetTerminal(Objects, 1)
	if err != nil {
		LOG.Println("Process power:", err)
		return
	}
	if in_a != in_b {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

var tbmu sync.Mutex
var tick float64

func ProcessTimeBase(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	tbmu.Lock()
	(*o)["NextOutput"] = tick //float64(time.Now().Unix())
	tbmu.Unlock()
	o.AssignOutput(Objects, 0)
}

func ProcessXYscope(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
}

func ProcessTimeRange(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	var on = o.GetProperty("on")
	var off = o.GetProperty("off")
	var timezone = stringify(o.GetProperty("timezone"))
	var loc, err = time.LoadLocation(timezone)
	if err != nil {
		(*o)["NextOutput"] = float64(0)
		o.AssignOutput(Objects, 0)
		return
	}
	var current_time = time.Now()
	var year, month, day = current_time.Date()
	m := int(month)
	on_time, err := time.ParseInLocation("15:04", on.(string), loc)
	if err != nil {
		(*o)["NextOutput"] = float64(0)
		o.AssignOutput(Objects, 0)
		return
	}
	on_time = on_time.AddDate(year, m-1, day-1)
	off_time, err := time.ParseInLocation("15:04", off.(string), loc)
	if err != nil {
		(*o)["NextOutput"] = float64(0)
		o.AssignOutput(Objects, 0)
		return
	}
	off_time = off_time.AddDate(year, m-1, day-1)

	if current_time.After(on_time) && current_time.Before(off_time) {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 0)
}

func ProcessTimer(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}
	var start int64
	var ok bool
	_, ok = (*o)["_timer_start"]
	if ok {
		start = (*o)["_timer_start"].(int64)
	}
	if !ok {
		start = time.Now().UTC().Unix()
		(*o)["_timer_start"] = start
	}
	now := time.Now().UTC().Unix()
	on := o.GetProperty("on duration")
	off := o.GetProperty("off duration")
	on_dur, err := time.ParseDuration(stringify(on))
	if err != nil {
		fmt.Println(err)
		return
	}
	off_dur, err := time.ParseDuration(stringify(off))
	if err != nil {
		fmt.Println(err)
		return
	}
	on_secs := int64(on_dur / time.Second)
	off_secs := int64(off_dur / time.Second)
	//fmt.Println("on_secs", on_secs)
	//fmt.Println("off_secs", off_secs)
	modsecs := (now - start) % (on_secs + off_secs)
	if modsecs >= 0 && modsecs < on_secs {
		(*o)["NextOutput"] = float64(1)
	} else if modsecs >= on_secs {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 0)
}

func ProcessDelay(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
	delay, ok := o.GetProperty("delay").(float64)
	if !ok {
		delay = float64(0)
	}
	min, ok := o.GetProperty("min on").(float64)
	if !ok {
		min = float64(0)
	}
	input, err := o.GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process delay:", err)
		return
	}
	current_time := time.Now()
	current_delay_start, ok := (*o)["_current_delay"].(time.Time)
	if !ok {
		current_delay_start = current_time
		if input > 0 && delay > 0 {
			(*o)["_current_delay"] = current_delay_start
			return
		}
	}
	delay_end := current_delay_start.Add(time.Duration(delay) * time.Second)

	if current_time.Equal(delay_end) || current_time.After(delay_end) {
		current_min_start, ok := (*o)["_current_min_on"].(time.Time)
		if !ok {
			current_min_start = current_time
			if min > 0 {
				(*o)["_current_min_on"] = current_min_start
			}
		}
		min_end := current_min_start.Add(time.Duration(min) * time.Second)

		if current_time.Equal(min_end) || current_time.After(min_end) {
			if input > 0 {
				(*o)["NextOutput"] = float64(1)
			} else {
				(*o)["NextOutput"] = float64(0)
				(*o)["_current_delay"] = nil
				(*o)["_current_min_on"] = nil
			}
		} else {
			(*o)["NextOutput"] = float64(1)
		}
	} else {
		if !(input > 0) {
			(*o)["_current_delay"] = nil
			(*o)["_current_min_on"] = nil
		}
		(*o)["NextOutput"] = float64(0)
	}

	o.AssignOutput(Objects, 1)
}

func ProcessConversion(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(2) {
		return
	}
	a, ok := o.GetProperty("a").(float64)
	if !ok {
		a = 0
	}
	b, ok := o.GetProperty("b").(float64)
	if !ok {
		b = 0
	}
	c, ok := o.GetProperty("c").(float64)
	if !ok {
		c = 0
	}
	input, err := (*o).GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process conversion:", err)
		return
	}
	(*o)["NextOutput"] = a*(input*input) + b*input + c
	o.AssignOutput(Objects, 1)
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

func ProcessLogger(o *Object_t, Objects map[int]*Object_t, iteration int) {
	if o.CheckTerminals(1) {
		return
	}

	input, err := (*o).GetTerminal(Objects, 0)
	if err != nil {
		LOG.Println("Process logger:", err)
		return
	}
	objid_lookup, ok := (*o)["Id"]
	if !ok {
		LOG.Println("Process logger: Object has no Id property")
	}
	objid, ok := objid_lookup.(int)
	if !ok {
		LOG.Println("Proccess logger: Object Id is not the correct type.")
		return
	}
	min, ok := (*o)["_min_value"].(float64)
	if !ok {
		min = input
		(*o)["_min_value"] = min
	}
	max, ok := (*o)["_max_value"].(float64)
	if !ok {
		max = input
		(*o)["_max_value"] = max
	}
	avg, ok := (*o)["_avg_data"].([]float64)
	if !ok {
		avg = []float64{}
		(*o)["_avg_data"] = avg
	}
	freq, ok := o.GetProperty("frequency").(float64)
	if !ok {
		freq = 300
	}

	current := time.Now()

	next_timeslot, ok := (*o)["_next_timeslot"].(time.Time)
	if !ok {
		_, next_timeslot = getSurroundingTimeslots(current, freq)
		(*o)["_next_timeslot"] = next_timeslot
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
		(*o)["_avg_data"] = avg
		(*o)["_min_value"] = min
		(*o)["_max_value"] = max
		(*o)["_next_timeslot"] = next_timeslot
	}
	if input < min {
		min = input
		(*o)["_min_value"] = min
	}
	if input > max {
		max = input
		(*o)["_max_value"] = max
	}
	if iteration == 0 {
		avg = append(avg, input)
		(*o)["_avg_data"] = avg
	}
	(*o)["NextOutput"] = input
}
