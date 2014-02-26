package logic_engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"os"
	"timecl/app/network_manager"
)

func (p processor) MarshalJSON() ([]byte, error) {
	return []byte("[]"), nil
}

func (p *processor) GobEncode() ([]byte, error) {
	return []byte(""), nil
}
func (p *processor) GobDecode([]byte) error {
	return nil
}

func (e *Engine_t) Save() {
	LOG.Println("Saving")
	path, found := revel.Config.String("engine.savefile")
	if !found {
		return
	}
	tmp := make([]interface{}, 0)
	gob.Register(tmp)
	var p processor
	gob.Register(p)
	gob.Register(network_manager.PortURI{})
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(e)
	if err != nil {
		LOG.Println("Encoding:", err)
		return
	}
	err = ioutil.WriteFile(path+".new", m.Bytes(), 0600)
	if err != nil {
		LOG.Println(err)
		return
	}
	if _, err = os.Stat(path); err == nil {
		// main path exists
		if _, err = os.Stat(path + ".save"); err == nil {
			// backup exists
			err = os.Remove(path + ".save")
			if err != nil {
				LOG.Println(err)
				return
			}
		}
		err = os.Link(path, path+".save")
		if err != nil {
			LOG.Println(err)
			return
		}
		err = os.Remove(path)
		if err != nil {
			LOG.Println(err)
			return
		}
	}
	err = os.Link(path+".new", path)
	if err != nil {
		LOG.Println(err)
		return
	}
	err = os.Remove(path + ".new")
	if err != nil {
		LOG.Println(err)
		return
	}
	err = os.Remove(path + ".save")
	if err != nil {
		LOG.Println(err)
		return
	}
}

func (e *Engine_t) ReadAndDecode(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return
	}
	n, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	tmp := make([]interface{}, 0)
	gob.Register(tmp)
	var proc processor
	gob.Register(proc)
	gob.Register(network_manager.PortURI{})
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)

	err = dec.Decode(e)
	if err != nil {
		return
	}
	return nil
}

func (e *Engine_t) LoadObjects() {
	path, found := revel.Config.String("engine.savefile")
	if !found {
		fmt.Println("No save file in configuration.")
		return
	}
	err := e.ReadAndDecode(path)
	if err != nil {
		LOG.Println(err)
		eagain := e.ReadAndDecode(path + ".save")
		if eagain != nil {
			LOG.Println(eagain)
			return
		}
	}
	for k, _ := range e.Objects {
		obj := *e.Objects[k]
		obj["process"] = processors[obj["Type"].(string)]
	}
}
