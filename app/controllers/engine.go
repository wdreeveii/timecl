package controllers

import (
	"code.google.com/p/go.net/websocket"
	"github.com/revel/revel"
	"timecl/app/logic_engine"
)

var (
	engine *logic_engine.Engine_t
)

type Engine struct {
	Application
}

func init() {
	revel.OnAppStart(InitEngine)
}

func InitEngine() {
	engine = &logic_engine.Engine_t{}
	engine.Init()
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

func (c Engine) NewEngine() revel.Result {
	InitEngine()
	engine.Save()
	return c.RenderJson(1)
}

func (c Engine) EngineSocket(ws *websocket.Conn) revel.Result {
	subscription := engine.Subscribe()
	defer subscription.Cancel()
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
	revel.INFO.Println(ports)

	newMessages := make(chan logic_engine.Event)
	go func() {
		var msg logic_engine.Event
		for {
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				revel.INFO.Println("Error receiving msg from client:", err)
				close(newMessages)
				return
			}
			newMessages <- msg
		}
	}()

	for {
		select {
		case event := <-subscription.New:
			if err := websocket.JSON.Send(ws, &event); err != nil {
				revel.INFO.Println("Error sending msg to client:", err)
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
