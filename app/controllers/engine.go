package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"time"
	"timecl/app/logger"
	"timecl/app/logic_engine"
	"timecl/app/models"
	"timecl/app/routes"
)

var (
	engine *logic_engine.Engine_t
)

type Engine struct {
	Application
}

func InitEngine(dbm *gorp.DbMap) {
	dataBasePath, found := revel.Config.String("engine.datadir")
	if !found {
		revel.ERROR.Fatal("No engine data file path provided in config file.")
	}

	path := dataBasePath
	txn, err := dbm.Begin()
	if err != nil {
		path += "/default.logic"
		logger.PublishOneError(fmt.Errorf("Problem initiating database transaction: %s", err))
	} else {
		instances, err := models.GetActiveEngineInstances(txn)
		txn.Commit()
		if err != nil {
			path += "/default.logic"
			logger.PublishOneError(fmt.Errorf("Problem finding active engine instances: %s", err))
		} else if len(instances) == 0 {
			path += "/default.logic"
		} else {
			path += "/" + instances[0].DataFile + ".logic"
		}
	}
	engine = logic_engine.Init(path)
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
	var status bool
	err = c.Txn.SelectOne(&status, "SELECT Enabled FROM EngineInstance WHERE EngineInstance.Id = ?", id)
	if err != nil {
		return c.RenderError(err)
	}
	return c.RenderJson(status)
}

func (c Engine) SaveEngine(engine_info models.EngineInstance) revel.Result {
	engine_info.Validate(c.Validation)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.Engine.CreateNewEngine())
	}
	err := models.SaveEngineInstance(c.Txn, engine_info)
	if err != nil {
		c.RenderError(err)
	}

	//InitEngine(dataPath)
	//engine.Stop()
	return c.Redirect(routes.Admin.SystemSettings())
}

func setDeadline(ws *websocket.Conn) error {
	return ws.SetDeadline(time.Now().Add(60 * time.Second))
}
func (c Engine) EngineSocket(ws *websocket.Conn) revel.Result {
	// Close transaction to prevent database writer starvation
	c.Commit()

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
			fmt.Println("msg from client")
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
		case msg, ok := <-newMessages:
			if !ok {
				return nil
			}
			engine.Publish(msg)
		}
	}
	return nil
}
