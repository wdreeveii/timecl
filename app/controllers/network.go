package controllers

import (
	"fmt"
	//"sort"
	"github.com/revel/revel"
	"timecl/app/models"
	"timecl/app/network_manager"
	"timecl/app/routes"
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

func (c Network) ShowDevices() revel.Result {
	network_definition := network_manager.ListPorts()
	return c.Render(network_definition)
}

func (c Network) EditNetwork(NetworkID int) revel.Result {
	results, err := c.Txn.Select(network_manager.NetworkConfig{}, `select * from NetworkConfig where NetworkID = ?`, NetworkID)
	if err != nil {
		panic(err)
	}
	if len(results) > 0 {
		network := results[0].(*network_manager.NetworkConfig)
		available_drivers := network_manager.GetDriverList()
		return c.Render(network, available_drivers)
	} else {
		return c.NotFound("Could not find network.")
	}
}

func (c Network) SaveNetwork(NetworkID int, network network_manager.NetworkConfig) revel.Result {
	network.NetworkID = NetworkID
	_, err := c.Txn.Update(&network)
	if err != nil {
		panic(err)
	}
	network_manager.RestartDriver(network.NetworkID, network.Driver)
	return c.Redirect(routes.Network.EditNetwork(NetworkID))
}
