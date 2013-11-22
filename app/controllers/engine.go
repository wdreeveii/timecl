package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	//"sort"
	"github.com/robfig/revel"
	"timecl/app/logic_engine"
	//"timecl/app/network_manager"
	//"timecl/app/routes"
	//"strings"
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
	fmt.Println("Engine checkuser")
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
		fmt.Println(err)
		return nil
	}
	fmt.Println("Websocket")
	ports := engine.ListPorts()
	if err := websocket.JSON.Send(ws, &ports); err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(ports)
	/*for _, event := range subscription.Archive {
		if websocket.JSON.Send(ws, &event) != nil {
			return nil
		}
	}*/

	newMessages := make(chan logic_engine.Event)
	go func() {
		var msg logic_engine.Event
		for {
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				fmt.Println(err)
				close(newMessages)
				return
			}
			fmt.Println("newMessage recv")
			newMessages <- msg
		}
	}()

	for {
		select {
		case event := <-subscription.New:
			if err := websocket.JSON.Send(ws, &event); err != nil {
				fmt.Println("jsoning", err)
				//return nil
			}
		case msg, ok := <-newMessages:
			if !ok {
				fmt.Println("recving", ok)
				return nil
			}
			//fmt.Println("new msg:", msg)
			engine.Publish(msg)
		}
	}
	return nil
}
