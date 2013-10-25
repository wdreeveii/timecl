package logic_engine

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var processors = make(map[string]processor)

func init() {
	processors["guide"] = ProcessGuide
	processors["binput"] = ProcessBinput
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
	processors["xyscope"] = ProcessXYscope
	//processors["block"] = 
	//processors["vbar"] = 
	//processors["hbar"] = 
	processors["timebase"] = ProcessTimeBase
	processors["timerange"] = ProcessTimeRange
	processors["timer"] = ProcessTimer
	go func() {
		for {
			<-time.After(1000 * time.Millisecond)
			tbmu.Lock()
			tick += 1
			tbmu.Unlock()
		}
	}()

}

func ProcessGuide(o *Object_t, Objects map[int]*Object_t) {
	source := int((*o)["Source"].(int))
	if source < 0 {
		return
	}
	(*o)["NextOutput"] = (*Objects[source])["Output"]
}

func ProcessBinput(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(1) {
		return
	}
	(*o)["NextOutput"] = (*o)["Output"]
	term := int((*o)["Terminals"].([]interface{})[0].(float64))
	(*Objects[term])["NextOutput"] = (*o)["Output"]
}

func ProcessBoutput(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(1) {
		return
	}
	term := int((*o)["Terminals"].([]interface{})[0].(float64))
	(*o)["NextOutput"] = (*Objects[term])["Output"]
}

func ProcessAoutput(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(1) {
		return
	}
	term := int((*o)["Terminals"].([]interface{})[0].(float64))
	(*o)["NextOutput"] = (*Objects[term])["Output"]
}

func ProcessNotGate(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(2) {
		return
	}
	if o.GetTerminal(Objects, 0) > 0 {
		(*o)["NextOutput"] = float64(0)
	} else {
		(*o)["NextOutput"] = float64(1)
	}
	o.AssignOutput(Objects, 1)
}

func ProcessAndGate(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	if (*Objects[term0])["Output"].(float64) > 0 && (*Objects[term1])["Output"].(float64) > 0 {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessOrGate(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	if (*Objects[term0])["Output"].(float64) > 0 || (*Objects[term1])["Output"].(float64) > 0 {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func xor(cond1, cond2 bool) bool {
	return (cond1 || cond2) && !(cond1 && cond2)
}

func ProcessXorGate(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	if xor(((*Objects[term0])["Output"].(float64) > 0), ((*Objects[term1])["Output"].(float64) > 0)) {
		(*o)["NextOutput"] = float64(1)
	} else {
		(*o)["NextOutput"] = float64(0)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessMult(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	(*o)["NextOutput"] = (*Objects[term0])["Output"].(float64) * (*Objects[term1])["Output"].(float64)
	o.AssignOutput(Objects, 2)
}

func ProcessDiv(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	if (*Objects[term1])["Output"].(float64) != 0 {
		(*o)["NextOutput"] = (*Objects[term0])["Output"].(float64) / (*Objects[term1])["Output"].(float64)
	}
	o.AssignOutput(Objects, 2)
}

func ProcessAdd(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	(*o)["NextOutput"] = (*Objects[term0])["Output"].(float64) + (*Objects[term1])["Output"].(float64)
	o.AssignOutput(Objects, 2)
}

func ProcessSub(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	(*o)["NextOutput"] = (*Objects[term0])["Output"].(float64) - (*Objects[term1])["Output"].(float64)
	o.AssignOutput(Objects, 2)
}

func ProcessPower(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(3) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	term1 := int((*o)["Terminals"].([]interface{})[1].(float64))
	(*o)["NextOutput"] = math.Pow((*Objects[term0])["Output"].(float64), (*Objects[term1])["Output"].(float64))
	o.AssignOutput(Objects, 2)
}

func ProcessSine(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(2) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	(*o)["NextOutput"] = math.Sin((*Objects[term0])["Output"].(float64))
	o.AssignOutput(Objects, 1)
}

func ProcessCosine(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(2) {
		return
	}
	term0 := int((*o)["Terminals"].([]interface{})[0].(float64))
	(*o)["NextOutput"] = math.Cos((*Objects[term0])["Output"].(float64))
	o.AssignOutput(Objects, 1)
}

var tbmu sync.Mutex
var tick float64

func ProcessTimeBase(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(1) {
		return
	}
	tbmu.Lock()
	(*o)["NextOutput"] = tick //float64(time.Now().Unix())
	tbmu.Unlock()
	o.AssignOutput(Objects, 0)
}

func ProcessXYscope(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(2) {
		return
	}
}

/*			if ($this->check_terminals(1)) return;

			$on = $this->get_property("on");
			$off = $this->get_property("off");

			$on_time = strtotime($on);
			$off_time = strtotime($off);

			$current = time();

			if ($current >= $on_time && $current <= $off_time)
				$this->output = 1;
			else
				$this->output = 0;


			$this->next_output =  $this->output;

			$Objects[$this->terminals[0]]->next_output = $this->output;*/
func ProcessTimeRange(o *Object_t, Objects map[int]*Object_t) {
	if o.CheckTerminals(1) {
		return
	}
	var on = o.GetProperty("on")
	var off = o.GetProperty("off")
	var timezone = stringify(o.GetProperty("timezone"))
	var loc, err = time.LoadLocation(timezone)
	if err != nil {
		return
	}
	var current_time = time.Now().UTC()
	var year, month, day = current_time.Date()
	m := int(month)
	var on_time, _ = time.ParseInLocation("15:04", on.(string), loc)
	on_time = on_time.AddDate(year, m-1, day-1)
	var off_time, _ = time.ParseInLocation("15:04", off.(string), loc)
	off_time = off_time.AddDate(year, m-1, day-1)

	if current_time.After(on_time) && current_time.Before(off_time) {
		(*o)["Output"] = float64(1)
	} else {
		(*o)["Output"] = float64(0)
	}
	(*o)["NextOutput"] = (*o)["Output"]
	term0 := intify((*o)["Terminals"].([]interface{})[0])
	(*Objects[term0])["NextOutput"] = (*o)["Output"]
}

func ProcessTimer(o *Object_t, Objects map[int]*Object_t) {
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
	fmt.Println("on_secs", on_secs)
	fmt.Println("off_secs", off_secs)
	modsecs := (now - start) % (on_secs + off_secs)
	if modsecs >= 0 && modsecs < on_secs {
		(*o)["Output"] = float64(1)
	} else if modsecs >= on_secs {
		(*o)["Output"] = float64(0)
	}
	(*o)["NextOutput"] = (*o)["Output"]
	term0 := intify((*o)["Terminals"].([]interface{})[0])
	(*Objects[term0])["NextOutput"] = (*o)["Output"]
}
