package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	//"sort"
	"github.com/robfig/revel"
	"time"
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
func (c Engine) Show() revel.Result {
	return c.Render()
}

func (c Engine) ListObjects() revel.Result {
	var objects []logic_engine.Object_t
	objects = engine.ListObjects()
	fmt.Println(objects)
	return c.RenderJson(objects)
}

func (c Engine) GetStates(state int) revel.Result {
	time.Sleep(40 * time.Second)
	states := engine.GetStates()
	return c.RenderJson(states)
}
func (c Engine) SetOutput(id int, output float64) revel.Result {
	engine.SetOutput(id, output)
	return c.RenderJson(1)
}
func (c Engine) SetProperties(id int, property_count int,
	property_names []string, property_types []string, property_values []string) revel.Result {
	engine.SetProperties(id, property_count, property_names, property_types, property_values)
	return c.RenderJson(1)
}
func (c Engine) HookObject(id int, source int) revel.Result {
	engine.HookObject(id, source)
	return c.RenderJson(1)
}
func (c Engine) UnhookObject(id int) revel.Result {
	engine.UnhookObject(id)
	return c.RenderJson(1)
}

func (c Engine) DeleteObject(id int) revel.Result {
	engine.DeleteObject(id)
	return c.RenderJson(1)
}

func (c Engine) MoveObject(id int, x_pos int, y_pos int) revel.Result {
	fmt.Println("Moving")
	engine.MoveObject(id, x_pos, y_pos)
	return c.RenderJson(1)
}

func (c Engine) NewEngine() revel.Result {
	InitEngine()
	engine.Save()
	return c.RenderJson(1)
}

func (c Engine) EngineSocket(ws *websocket.Conn) revel.Result {
	subscription := engine.Subscribe()
	defer subscription.Cancel()

	for _, event := range subscription.Archive {
		if websocket.JSON.Send(ws, &event) != nil {
			return nil
		}
	}

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
			fmt.Println("new msg:", msg)
			engine.Publish(msg)
		}
	}
	return nil
}
