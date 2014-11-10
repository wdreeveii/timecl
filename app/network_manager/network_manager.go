package network_manager

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"sort"
	"time"
	"github.com/wdreeveii/timecl/app/logger"
)

type DriverInterface interface {
	// Called on server startup
	Init(port string, network_id int)
	Stop()
	ListPorts() []BusDef
}

type driverListItem struct {
	Name     string
	Instance DriverInterface
}

var (
	driver_collection []driverListItem
)

func GetDriverList() []string {
	var drivers []string
	for _, val := range driver_collection {
		drivers = append(drivers, val.Name)
	}
	return drivers
}

func RegisterDriver(drivername string, driver DriverInterface) {
	driver_collection = append(driver_collection, driverListItem{Name: drivername, Instance: driver})
}

type interfaceItem struct {
	Driver    driverListItem
	ConfigKey string
}

type restartConfig struct {
	NetworkID int
	Driver    string
}

var (
	restartInterface = make(chan restartConfig)
	newInterface     = make(chan interfaceItem)
	subscribe        = make(chan SubscriptionRequest, 10)
	unsubscribe      = make(chan (<-chan Event), 10)
	publish          = make(chan Event, 100)
	list_ports       = make(chan (chan []NetInterfaceDef))
)

type EventArgument interface{}

type Event struct {
	NetworkID int
	Type      string
	Data      EventArgument
	Timestamp int
}
type SubscriptionRequest struct {
	NetworkID     int
	FilterNetwork bool
	Types         []string
	FilterTypes   bool
	Response      chan Subscription
}

type Subscription struct {
	NetworkID     int
	FilterNetwork bool
	Types         []string
	FilterTypes   bool
	New           <-chan Event
	newSendable   chan Event
}

func (s Subscription) Cancel() {
	unsubscribe <- s.New
	drain(s.New)
}

func subscribeBase(netid int,
	filter_net bool,
	types []string,
	filter_type bool) Subscription {
	var req SubscriptionRequest
	req.Response = make(chan Subscription)
	req.NetworkID = netid
	req.FilterNetwork = filter_net
	req.Types = types
	req.FilterTypes = filter_type
	subscribe <- req
	return <-req.Response
}

func Subscribe() Subscription {
	return subscribeBase(0, false, []string{}, false)
}
func SubscribeNetwork(NetworkID int) Subscription {
	return subscribeBase(NetworkID, true, []string{}, false)
}
func SubscribeType(Types []string) Subscription {
	return subscribeBase(0, false, Types, true)
}
func SubscribeNetworkTypes(NetworkID int, Types []string) Subscription {
	return subscribeBase(NetworkID, true, Types, true)
}

type PortURI struct {
	Network int
	Bus     int
	Device  int
	Port    int
}

type Value_t float64

type PortChange struct {
	URI   PortURI
	Value Value_t
}

type GetData struct {
	BusID    int
	DeviceID int
	PortID   int
	Recv     chan float64
}

func Get(port PortURI) (float64, error) {
	var m = make(chan float64)
	Publish(NewEvent(port.Network, "get", GetData{BusID: port.Bus, DeviceID: port.Device, PortID: port.Port, Recv: m}))
	select {
	case newval, ok := <-m:
		if !ok {
			return 0, errors.New("Driver error.")
		}
		return newval, nil
	case <-time.After(20 * time.Millisecond):
		return 0, errors.New("Get Time Out")
	}
	return 0, errors.New("Unreachable.")
}

type SetData struct {
	BusID    int
	DeviceID int
	PortID   int
	Value    Value_t
}

func PublishSetEvents(in []PortChange) {
	var sorted_events = make(map[int][]PortChange)
	for _, v := range in {
		sorted_events[v.URI.Network] = append(sorted_events[v.URI.Network], v)
	}
	for network, v := range sorted_events {
		var sd []SetData
		for _, data := range v {
			sd = append(sd, SetData{BusID: data.URI.Bus, DeviceID: data.URI.Device, PortID: data.URI.Port, Value: data.Value})
		}
		Publish(NewEvent(network, "set", sd))
	}
}

func NewEvent(net_id int, typ string, data EventArgument) Event {
	return Event{net_id, typ, data, int(time.Now().Unix())}
}

func Publish(event Event) {
	publish <- event
}

//get (network number, bus number, device number, port number
func RestartDriver(NetworkID int, driver string) {
	restartInterface <- restartConfig{NetworkID: NetworkID, Driver: driver}
}

type PortFunction int

const (
	BInput PortFunction = iota
	AInput
	BOutput
	AOutput
)

func (t PortFunction) String() string {
	switch t {
	case BInput:
		return "BInput"
	case AInput:
		return "AInput"
	case BOutput:
		return "BOutput"
	case AOutput:
		return "AOutput"
	}
	return ""
}

type PortDef struct {
	PortID uint32
	Type   PortFunction
}
type DeviceDef struct {
	DeviceID uint32
	PortList []PortDef
}
type BusDef struct {
	BusID      uint32
	DeviceList []DeviceDef
}
type NetInterfaceDef struct {
	NetworkID uint32
	BusList   []BusDef
}

func ListPorts() []NetInterfaceDef {
	var m = make(chan []NetInterfaceDef)
	list_ports <- m
	return <-m
}

func interfacesManager() {
	var interfaces []interfaceItem
	subscribers := list.New()
	for {
		select {
		case req := <-list_ports:
			var res = make([]NetInterfaceDef, 0)
			for idx, aInterface := range interfaces {
				if aInterface.Driver.Instance != nil {
					var item NetInterfaceDef
					item.NetworkID = uint32(idx)
					item.BusList = aInterface.Driver.Instance.ListPorts()
					res = append(res, item)
				}
			}
			req <- res
		case ch := <-subscribe:
			subscriber := make(chan Event, 100)
			sub := Subscription{NetworkID: ch.NetworkID,
				FilterNetwork: ch.FilterNetwork,
				Types:         ch.Types,
				FilterTypes:   ch.FilterTypes,
				New:           subscriber,
				newSendable:   subscriber}
			ch.Response <- sub
			subscribers.PushBack(sub)
		case event := <-publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				subscription := ch.Value.(Subscription)
				var network_match bool = false
				if subscription.FilterNetwork {
					if event.NetworkID == subscription.NetworkID {
						network_match = true
					}
				} else {
					network_match = true
				}
				var type_match bool = false
				if subscription.FilterTypes {
					if sort.SearchStrings(subscription.Types, event.Type) == len(subscription.Types) {
						type_match = true
					}
				} else {
					type_match = true
				}
				if network_match == true && type_match == true {
					subscription.newSendable <- event
				}
			}
		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(Subscription).New == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		case reConfig := <-restartInterface:
			if reConfig.NetworkID < len(interfaces) {
				var iface = interfaces[reConfig.NetworkID]
				if iface.Driver.Instance != nil {
					iface.Driver.Instance.Stop()
				}
				for _, a_driver := range driver_collection {
					if reConfig.Driver == a_driver.Name {
						interfaces[reConfig.NetworkID].Driver = a_driver
						device_path, found := revel.Config.String(iface.ConfigKey)
						if found {
							interfaces[reConfig.NetworkID].Driver.Instance.Init(device_path, reConfig.NetworkID)
						}
						break
					}
				}
				logger.PublishOneError(fmt.Errorf("Driver not found for %s", iface.ConfigKey))
			}
		case newIntConfig := <-newInterface:
			if newIntConfig.Driver.Name != "" {
				val, found := revel.Config.String(newIntConfig.ConfigKey)
				if found {
					newIntConfig.Driver.Instance.Init(val, len(interfaces))
					interfaces = append(interfaces, newIntConfig)
				}
			}
		}
	}
}

type NetworkConfig struct {
	NetworkID  int
	ConfigKey  string
	DevicePath string
	Driver     string
}

func (n *NetworkConfig) String() string {
	return fmt.Sprintf("Network Config (%d, %s, %s, %s)", n.NetworkID, n.ConfigKey, n.DevicePath, n.Driver)
}

func GetHardwareInterfaces() []string {
	result := revel.Config.Options("hardware.")
	sort.Strings(result)
	return result
}

func GetNetworks(txn *gorp.Transaction) []*NetworkConfig {
	interfaces := GetHardwareInterfaces()
	var networks []*NetworkConfig
	for index, val := range interfaces {
		var config NetworkConfig
		err := txn.SelectOne(&config, `select * from NetworkConfig where ConfigKey = ?`, val)
		if err != nil {
			path, found := revel.Config.String(val)
			if !found {
				logger.PublishOneError(fmt.Errorf("Device path for %s not found.", val))
				continue
			}
			_, err = txn.Exec("INSERT INTO NetworkConfig VALUES(?,?,?,?)", index, val, path, "")
			if err != nil {
				logger.PublishOneError(fmt.Errorf("Problem initializing interface %s: %s", val, err))
			}
			net := NetworkConfig{index, val, path, ""}
			_, err = txn.Update(&net)
			if err != nil {
				logger.PublishOneError(fmt.Errorf("Problem initializing interface %s: %s", val, err))
			}
			networks = append(networks, &net)
		} else {
			networks = append(networks, &config)
		}
	}
	return networks
}

func InitNetworkConfigTables(dbm *gorp.DbMap) {
	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}
	t := dbm.AddTable(NetworkConfig{}).SetKeys(true, "NetworkID")
	setColumnSizes(t, map[string]int{
		"ConfigKey":  100,
		"DevicePath": 1000,
		"Driver":     100,
	})
}

func Init(dbm *gorp.DbMap) {
	revel.INFO.Println("Network Manager Start")
	go interfacesManager()

	InitNetworkConfigTables(dbm)
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}
	var networks []*NetworkConfig
	txn, err := dbm.Begin()
	if err != nil {
		logger.PublishOneError(fmt.Errorf("Could not create database transaction: %s", err))
	} else {
		networks = GetNetworks(txn)
		if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
			logger.PublishOneError(fmt.Errorf("Could not commit db transaction: %s", err))
		}
	}
Interfaces:
	for _, net := range networks {
		for _, a_driver := range driver_collection {
			if net.Driver == a_driver.Name {
				newInterface <- interfaceItem{ConfigKey: net.ConfigKey, Driver: a_driver}
				continue Interfaces
			}
		}
		logger.PublishOneError(fmt.Errorf("Driver for %s not found.", net.ConfigKey))
	}
}

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
