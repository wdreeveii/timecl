package controllers

import "github.com/robfig/revel"

func init() {
	revel.RegisterPlugin(GorpPlugin{})
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod(Application.AddUser, revel.BEFORE)
	revel.InterceptMethod(Admin.checkUser, revel.BEFORE)
	revel.InterceptMethod(Network.checkUser, revel.BEFORE)
	revel.InterceptMethod(Hotels.checkUser, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)
}
