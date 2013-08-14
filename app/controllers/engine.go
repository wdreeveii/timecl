
package controllers

import (
	"fmt"
	//"sort"
	"time"
	"github.com/robfig/revel"
	"timecl/app/models"
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
	user_results, err := c.Txn.Select(models.User{}, `select * from User`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("results: %#v\n", user_results)
	var users []*models.User
	for _, r := range user_results {
		users = append(users, r.(*models.User))
	}
	fmt.Printf("users: %#v\n", users)
	return c.Render(users)
}
func (c Engine) Show() revel.Result {
	return c.Render()
}

func (c Engine) ListObjects() revel.Result {
	objects := engine.ListObjects()
	fmt.Println(objects)
	return c.RenderJson(objects)
}

func (c Engine) GetStates() revel.Result {
	time.Sleep(3 * time.Second)
	states := engine.GetStates()
	return c.RenderJson(states)
}
func (c Engine) SetOutput(id int, output float32) revel.Result {
	engine.SetOutput(id, output)
	return c.RenderJson(1)
}
func (c Engine) SetProperties() revel.Result {
	return c.Render()
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
func (c Engine) AddObject(objtype string,
							root_id int,
							x_pos int,
							y_pos int,
							x_size int,
							y_size int,
							attached int,
							dir int,
							property_count int,
							property_names string,
							property_types string,
							property_values string ) revel.Result {
			
	newobj_id := engine.AddObject(objtype, x_pos, y_pos,
										x_size, y_size,attached,dir,
										property_count,property_names,
										property_types, property_values)
	if root_id > -1 {
		engine.SetGuides(root_id, newobj_id)
	}
	return c.RenderJson(newobj_id)
}

func (c Engine) MoveObject(id int, x_pos int, y_pos int) revel.Result {
	fmt.Println("Moving")
	engine.MoveObject(id, x_pos, y_pos)
	return c.RenderJson(1)
}

