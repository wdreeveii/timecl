
package controllers

import (
	"fmt"
	"github.com/robfig/revel"
	"github.com/robfig/revel/samples/booking/app/models"
	//"strings"
)

type Network struct {
	Application
}

func (c Network) checkUser() revel.Result {
	fmt.Println("Network checkuser")
	if user := c.connected(); user == nil {
		c.Flash.Error("Please log in first")
		return c.Redirect(Application.Index)
	}
	return nil
}

func (c Network) Index() revel.Result {
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
