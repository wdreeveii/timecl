package logic_engine

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"timecl/app/logger"
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
	interface_slice := make([]interface{}, 0)
	gob.Register(interface_slice)
	var p processor
	gob.Register(p)
	gob.Register(network_manager.PortURI{})
	gob.Register(time.Time{})
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	err := enc.Encode(e)
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Save file encoding error:", err))
		return
	}
	err = ioutil.WriteFile(e.DataFile+".new", m.Bytes(), 0600)
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Save file write error:", err))
		return
	}
	if _, err = os.Stat(e.DataFile); err == nil {
		// main path exists
		if _, err = os.Stat(e.DataFile + ".save"); err == nil {
			// backup exists
			err = os.Remove(e.DataFile + ".save")
			if err != nil {
				logger.PublishOneError(fmt.Errorf("Backup save file removal error:", err))
				return
			}
		}
		err = os.Link(e.DataFile, e.DataFile+".save")
		if err != nil {
			logger.PublishOneError(fmt.Errorf("Error creating backup save file:", err))
			return
		}
		err = os.Remove(e.DataFile)
		if err != nil {
			logger.PublishOneError(fmt.Errorf("Error removing old save file:", err))
			return
		}
	}
	err = os.Link(e.DataFile+".new", e.DataFile)
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Error swapping original save file with new save file:", err))
		return
	}
	err = os.Remove(e.DataFile + ".new")
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Error removing new save file:", err))
		return
	}
	err = os.Remove(e.DataFile + ".save")
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Error removing backup save file:", err))
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
	gob.Register(time.Time{})
	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)

	err = dec.Decode(e)
	if err != nil {
		return
	}
	return nil
}

func (e *Engine_t) LoadObjects() error {
	if e.DataFile == "" {
		return errors.New("Load Engine: no source data path provided.")
	}
	path := e.DataFile
	err := e.ReadAndDecode(path)
	if err != nil {
		DEBUG.Println(err)
		eagain := e.ReadAndDecode(path + ".save")
		if eagain != nil {
			return fmt.Errorf("Load Engine: specified engine data source not found. %s", eagain)
		}
	}
	for k, _ := range e.Objects {
		e.Objects[k].process = processors[e.Objects[k].Type]
	}
	return nil
}
