package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/revel/revel"
	"time"
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
	engine = logic_engine.Init()
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
func setDeadline(ws *websocket.Conn) error {
	return ws.SetDeadline(time.Now().Add(60 * time.Second))
}
func (c Engine) EngineSocket(ws *websocket.Conn) revel.Result {
	// Close transaction to prevent database writer starvation
	c.Commit()

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
		case event := <-subscription.New:
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
