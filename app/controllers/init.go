package controllers

import "github.com/revel/revel"

func init() {
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*Application).AddUser, revel.BEFORE)
	revel.InterceptMethod(Admin.checkUser, revel.BEFORE)
	revel.InterceptMethod(Engine.checkUser, revel.BEFORE)
	revel.InterceptMethod(Network.checkUser, revel.BEFORE)
	revel.InterceptMethod(Logger.checkUser, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)
}
