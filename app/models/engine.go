package models

import (
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"os"
	"regexp"
	"strings"
	"time"
)

type EngineInstance struct {
	Id        int
	DataFile  string
	TsCreated int64
	Enabled   bool

	Created time.Time
}

func (e *EngineInstance) PreInsert(_ gorp.SqlExecutor) error {
	e.TsCreated = e.Created.Unix()
	return nil
}

func (e *EngineInstance) PreUpdate(_ gorp.SqlExecutor) error {
	e.TsCreated = e.Created.Unix()
	return nil
}

func (e *EngineInstance) PostGet(_ gorp.SqlExecutor) error {
	e.Created = time.Unix(e.TsCreated, 0)
	return nil
}

func (engine_info *EngineInstance) Validate(v *revel.Validation) {
	var msg = "The instance name must contain only alphanumeric digits and underscore"
	v.Required(engine_info.DataFile).Message(msg + " MSG 1")
	v.Match(engine_info.DataFile, regexp.MustCompile(`\w+`)).Message(msg + " MSG 2")
}

func GetActiveEngineInstances(txn *gorp.Transaction) ([]*EngineInstance, error) {
	var list []*EngineInstance
	_, err := txn.Select(&list, "SELECT * FROM EngineInstance WHERE EngineInstance.Enabled = true")
	return list, err
}

func getEngineDataFiles() ([]string, error) {
	basePath, found := revel.Config.String("engine.datadir")
	if !found {
		return nil, errors.New("No engine.datadir property set in config file.")
	}
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open engine.datadir path: %s", err)
	}
	if !info.IsDir() {
		return nil, errors.New("engine.datadir path is not a directory.")
	}
	f, err := os.Open(basePath)
	if err != nil {
		return nil, fmt.Errorf("Problem opening engine.datadir path: %s", err)
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("Problem getting a list of files in engine.datadir path %s", err)
	}

	var formatted_names = make([]string, 0, len(names))
	for _, v := range names {
		if strings.HasSuffix(v, ".logic") {
			formatted_names = append(formatted_names, strings.TrimSuffix(v, ".logic"))
		}
	}
	return formatted_names, nil
}

func GetRecognizedEngineInstances(txn *gorp.Transaction) ([]*EngineInstance, error) {
	data_files, err := getEngineDataFiles()
	if err != nil {
		return nil, err
	}
	var engine_instances []*EngineInstance
	_, err = txn.Select(&engine_instances, "SELECT * FROM EngineInstance")
	if err != nil {
		return nil, err
	}
	var dedup = make(map[string]bool)
	for _, v := range data_files {
		dedup[v] = false
	}
	for _, v := range engine_instances {
		dedup[v.DataFile] = true
	}
	for k, v := range dedup {
		if v == false {
			var e = &EngineInstance{DataFile: k, Enabled: false}
			err := SaveEngineInstance(txn, e)
			if err != nil {
				return nil, err
			}
		}
	}
	engine_instances = make([]*EngineInstance, 0, len(dedup))
	_, err = txn.Select(&engine_instances, "SELECT * FROM EngineInstance")
	if err != nil {
		return nil, err
	}
	return engine_instances, nil

}
func SaveEngineInstance(txn *gorp.Transaction, engine_info *EngineInstance) error {
	var err error
	if engine_info.TsCreated > 0 {
		engine_info.Created = time.Unix(engine_info.TsCreated, 0)
	}
	if engine_info.Id == 0 {
		err = txn.Insert(engine_info)
	} else {
		_, err = txn.Update(engine_info)
	}
	return err
}
