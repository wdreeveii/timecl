package controllers

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"time"
	"github.com/wdreeveii/timecl/app/logger"
	"github.com/wdreeveii/timecl/app/logic_engine"
	"github.com/wdreeveii/timecl/app/models"
	"github.com/wdreeveii/timecl/app/routes"
)

type Engine struct {
	Application
}

type useEngineRequest struct {
	Id        int
	Requester chan *logic_engine.Engine_t
}

type stopRequest struct {
	Id   int
	Done chan bool
}

type startRequest struct {
	Id       int
	DataFile string
}

var (
	startRequestChan     = make(chan startRequest)
	stopRequestChan      = make(chan stopRequest)
	useEngineRequestChan = make(chan useEngineRequest)
)

func ManageEngines(engineStore map[int]*logic_engine.Engine_t, basePath string) {
	for {
		select {
		case req := <-startRequestChan:
			_, exists := engineStore[req.Id]
			if exists {
				engineStore[req.Id].Stop()
			}
			engineStore[req.Id] = logic_engine.Init(basePath + "/" + req.DataFile + ".logic")
		case req := <-stopRequestChan:
			_, exists := engineStore[req.Id]
			if exists {
				engineStore[req.Id].Stop()
				delete(engineStore, req.Id)
			}
			req.Done <- true
		case req := <-useEngineRequestChan:
			_, exists := engineStore[req.Id]
			if exists {
				req.Requester <- engineStore[req.Id]
			} else {
				req.Requester <- nil
			}
		}
	}
}

func InitEngines(dbm *gorp.DbMap) {
	var engineStore = make(map[int]*logic_engine.Engine_t)
	dataBasePath, found := revel.Config.String("engine.datadir")
	if !found {
		revel.ERROR.Fatal("No engine data file path provided in config file.")
	}
	var defaultPath = dataBasePath + "/default.logic"
	txn, err := dbm.Begin()
	if err != nil {
		engineStore[0] = logic_engine.Init(defaultPath)
		logger.PublishOneError(fmt.Errorf("Problem initiating database transaction: %s", err))
	} else {
		instances, err := models.GetActiveEngineInstances(txn)
		txn.Commit()
		if err != nil {
			engineStore[0] = logic_engine.Init(defaultPath)
			logger.PublishOneError(fmt.Errorf("Problem finding active engine instances: %s", err))
		} else if len(instances) == 0 {
			engineStore[0] = logic_engine.Init(defaultPath)
		} else {
			for _, v := range instances {
				engineStore[v.Id] = logic_engine.Init(dataBasePath + "/" + v.DataFile + ".logic")
			}
		}
	}
	go ManageEngines(engineStore, dataBasePath)
}

func (c Engine) checkUser() revel.Result {
	if user := c.connected(); user == nil {
		c.Flash.Error("Please log in first")
		return c.Redirect(Application.Index)
	}
	return nil
}

func (c Engine) Index() revel.Result {
	return c.Render()
}

func (c Engine) CreateNewEngine() revel.Result {
	// get sample list
	var engine_info models.EngineInstance
	c.RenderArgs["engine_info"] = engine_info
	return c.RenderTemplate("Engine/AddEditEngine.html")
}

func (c Engine) EditEngine(id int) revel.Result {
	var engine_info models.EngineInstance
	err := c.Txn.SelectOne(&engine_info, "SELECT * FROM EngineInstance WHERE Id = ?", id)
	if err != nil {
		return c.RenderError(err)
	}
	c.RenderArgs["engine_info"] = engine_info
	return c.RenderTemplate("Engine/AddEditEngine.html")
}

func (c Engine) ToggleEngine(id int) revel.Result {
	_, err := c.Txn.Exec("UPDATE EngineInstance SET Enabled = !EngineInstance.Enabled WHERE EngineInstance.Id = ?", id)
	if err != nil {
		return c.RenderError(err)
	}
	var status struct {
		Enable   bool
		DataFile string
	}
	err = c.Txn.SelectOne(&status, "SELECT Enabled, DataFile FROM EngineInstance WHERE EngineInstance.Id = ?", id)
	if err != nil {
		return c.RenderError(err)
	}
	if status.Enable {
		startRequestChan <- startRequest{Id: id, DataFile: status.DataFile}
	} else {
		done := make(chan bool)
		stopRequestChan <- stopRequest{Id: id, Done: done}
		<-done
	}
	return c.RenderJson(status.Enable)
}

func (c Engine) SaveEngine(engine_info models.EngineInstance) revel.Result {
	engine_info.Validate(c.Validation)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		if engine_info.Id == 0 {
			return c.Redirect(routes.Engine.CreateNewEngine())
		} else {
			return c.Redirect(routes.Engine.EditEngine(engine_info.Id))
		}
	}
	err := models.SaveEngineInstance(c.Txn, &engine_info)
	if err != nil {
		c.RenderError(err)
	}
	if engine_info.Enabled {
		startRequestChan <- startRequest{Id: engine_info.Id, DataFile: engine_info.DataFile}
	} else {
		done := make(chan bool)
		stopRequestChan <- stopRequest{Id: engine_info.Id, Done: done}
		<-done
	}

	return c.Redirect(routes.Admin.SystemSettings())
}

func setDeadline(ws *websocket.Conn) error {
	return ws.SetDeadline(time.Now().Add(60 * time.Second))
}
func (c Engine) EngineSocket(ws *websocket.Conn, id int) revel.Result {
	// Close transaction to prevent database writer starvation
	c.Commit()

	logger_subscription := logger.Subscribe()
	defer logger_subscription.Cancel()

	response := make(chan *logic_engine.Engine_t)
	useEngineRequestChan <- useEngineRequest{Id: id, Requester: response}
	engine := <-response
	if engine == nil {
		return c.RenderError(errors.New("Requested engine not running."))
	}
	subscription := engine.Subscribe()
	defer engine.CancelSubscription(subscription)
	init := engine.ListObjects()
	if err := websocket.JSON.Send(ws, &init); err != nil {
		revel.INFO.Println(err)
		return nil
	}
	ports := engine.ListPorts()
	if err := websocket.JSON.Send(ws, &ports); err != nil {
		revel.INFO.Println(err)
		return nil
	}

	newMessages := make(chan logic_engine.Event)
	go func() {
		var msg logic_engine.Event
		for {
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				revel.ERROR.Println("Error receiving msg from client:", err)
				close(newMessages)
				return
			}
			if err := setDeadline(ws); err != nil {
				revel.ERROR.Println(err)
				close(newMessages)
				return
			}
			newMessages <- msg
		}
	}()

	for {
		select {
		case event, ok := <-subscription.New:
			if !ok {
				return nil
			}
			if err := websocket.JSON.Send(ws, &event); err != nil {
				revel.ERROR.Println("Error sending msg to client:", err)
				return nil
			}
			if err := setDeadline(ws); err != nil {
				revel.ERROR.Println(err)
				return nil
			}
		case event, ok := <-logger_subscription.New:
			if !ok {
				logger_subscription.New = nil
			}
			if event.Type == "errors" {
				if err := websocket.JSON.Send(ws, &event); err != nil {
					revel.ERROR.Println("Error sending msg to client:", err)
					return nil
				}
				if err := setDeadline(ws); err != nil {
					revel.ERROR.Println(err)
					return nil
				}
			}
		case msg, ok := <-newMessages:
			if !ok {
				return nil
			}
			engine.Publish(msg)
		}
	}
	return nil
}
