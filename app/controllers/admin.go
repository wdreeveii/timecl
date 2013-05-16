package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/robfig/revel"
	"github.com/robfig/revel/samples/booking/app/models"
	//"strings"
)

type Admin struct {
	Application
}

func (c Admin) checkUser() revel.Result {
	fmt.Println("admin checkuser")
	if user := c.connected(); user == nil {
		c.Flash.Error("Please log in first")
		return c.Redirect(Application.Index)
	}
	return nil
}

func (c Admin) Index() revel.Result {
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

func (c Admin) EditMyUser() revel.Result {
	//c.RenderArgs["user"] = user
	c.RenderArgs["id"] = c.connected().UserId
	return c.RenderTemplate("Admin/EditUser.html")
}

func (c Admin) EditUser(id int) revel.Result {
	return c.Render(id)
}

func (c Admin) SaveSettings(id int, password, verifyPassword string) revel.Result {
	models.ValidatePassword(c.Validation, password)
	c.Validation.Required(verifyPassword).
		Message("Please verify your password")
	c.Validation.Required(verifyPassword == password).
		Message("Your password doesn't match")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		return c.Redirect(Admin.EditUser)
	}

	bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := c.Txn.Exec("update User set HashedPassword = ? where UserId = ?",
		bcryptPassword, id)
	if err != nil {
		panic(err)
	}
	c.Flash.Success("Password updated")
	return c.Redirect(Admin.Index)
}

func (c Admin) AddUser() revel.Result {
	return c.Render()
}

func (c Admin) SaveUser(user models.User, verifyPassword string) revel.Result {
	c.Validation.Required(verifyPassword)
	c.Validation.Required(verifyPassword == user.Password).
		Message("Password does not match")
	user.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Admin.AddUser)
	}

	user.HashedPassword, _ = bcrypt.GenerateFromPassword(
		[]byte(user.Password), bcrypt.DefaultCost)
	err := c.Txn.Insert(&user)
	if err != nil {
		panic(err)
	}

	return c.Redirect(Admin.Index)
}
