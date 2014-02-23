package logic_engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/robfig/revel"
	"io/ioutil"
	"os"
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
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(e)
	if err != nil {
		LOG.Println("Encoding:", err)
	}
	err = ioutil.WriteFile(path, m.Bytes(), 0600)
	if err != nil {
		panic(err)
	}
}

func (e *Engine_t) LoadObjects() {
	path, found := revel.Config.String("engine.savefile")
	if !found {
		fmt.Println("No save file in configuration.")
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println(err)
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
		fmt.Println("Decode objects err:", err)
		return
	}
	for k, _ := range e.Objects {
		obj := *e.Objects[k]
		obj["process"] = processors[obj["Type"].(string)]
	}
}
