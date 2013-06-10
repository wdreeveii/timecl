
package network_manager

import (
	"fmt"
	"github.com/robfig/revel"
	"github.com/robfig/revel/modules/db/app"
	"github.com/coopernurse/gorp"
	"timecl/app/models"
	"sort"
)

type network_manager struct {
	*revel.Controller
}
type DriverInterface interface {
	// Called on server startup
	Init(port string)
	Stop()
	Get(cmd GetDrvCmd)
	Set(cmd SetDrvCmd)
}

type driverListItem struct {
	Name		string
	Instance	DriverInterface
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
	fmt.Println("REGISTER DRIVER: ", drivername)
	driver_collection = append(driver_collection, driverListItem{Name: drivername, Instance: driver})
}

func init_networkconfig_table(dbm *gorp.DbMap) {
	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}
	t := dbm.AddTable(models.NetworkConfig{}).SetKeys(true, "NetworkID")
	setColumnSizes(t, map[string]int{
		"ConfigKey":	100,
		"DevicePath":	1000,
		"Driver":		100,
	})
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}
}

func GetHardwareInterfaces() []string {
	result := revel.Config.Options("hardware.")
	sort.Strings(result)
	return result
}

type interfaceItem struct {
	Driver		driverListItem
	ConfigKey	string
}

type restartConfig struct {
	NetworkID	int
	Driver		string
}
type GetDrvCmd struct {
	Bus			int
	Device		int
	Port		int
	RecvChan	chan interface{}
}

type getCmd struct {
	NetworkID	int
	Cmd			GetDrvCmd
}

type SetDrvCmd struct {
	Bus			int
	Device		int
	Port		int
	Value		interface{}
}

type setCmd struct {
	NetworkID	int
	Cmd			SetDrvCmd
}

var restartInterface = make(chan restartConfig)
var newInterface = make(chan interfaceItem)
var getInterface = make(chan getCmd)
var setInterface = make(chan setCmd)

//get (network number, bus number, device number, port number
func Get(NetworkID int, bus int, device int, port int) interface{} {
	var cmd = getCmd{NetworkID: NetworkID, Cmd: GetDrvCmd{Bus: bus, Device: device, Port: port, RecvChan: make(chan interface{})}}
	getInterface <-cmd
	return <- cmd.Cmd.RecvChan
}

func Set(NetworkID int, bus int, device int, port int, value interface{}) {
	var cmd = setCmd{NetworkID: NetworkID, Cmd: SetDrvCmd{Bus: bus, Device: device, Port: port, Value: value}}
	setInterface <- cmd
}

func RestartDriver(NetworkID int, driver string) {
	restartInterface <- restartConfig{NetworkID: NetworkID, Driver: driver}
}

func interfacesManager() {
	var interfaces []interfaceItem
	for {
		var reConfig 		restartConfig
		var newIntConfig 	interfaceItem
		var getIntCmd		getCmd
		var setIntCmd		setCmd
		
		select {
		case getIntCmd = <- getInterface:
			interfaces[getIntCmd.NetworkID].Driver.Instance.Get(getIntCmd.Cmd)
		case setIntCmd = <- setInterface:
			interfaces[setIntCmd.NetworkID].Driver.Instance.Set(setIntCmd.Cmd)
		case reConfig = <- restartInterface:
			for _, val := range driver_collection {
				if val.Name == reConfig.Driver {
					if interfaces[reConfig.NetworkID].Driver.Name != "" {
						interfaces[reConfig.NetworkID].Driver.Instance.Stop()
					}
					interfaces[reConfig.NetworkID].Driver = val
					val, found := revel.Config.String(interfaces[reConfig.NetworkID].ConfigKey)
					if found {
						interfaces[reConfig.NetworkID].Driver.Instance.Init(val)
					}
				}
			}
		case newIntConfig = <- newInterface:
			val, found := revel.Config.String(newIntConfig.ConfigKey)
			if found {
				interfaces = append(interfaces, newIntConfig)
				if interfaces[len(interfaces)-1].Driver.Name != "" {
					interfaces[len(interfaces) - 1].Driver.Instance.Init(val)
				}
			}
		}
	}
}

func Init() {
	fmt.Println("driver start")
	go interfacesManager()
	db.Init()
	dbm := &gorp.DbMap{Db: db.Db, Dialect: gorp.SqliteDialect{}}

	init_networkconfig_table(dbm)
	
	result := GetHardwareInterfaces()
	fmt.Println("results: ", result)
	for _, config_key := range result {
		fmt.Println(config_key)
		networks, err := dbm.Select(models.NetworkConfig{}, `select * from NetworkConfig where ConfigKey = ?`, config_key)
		if err != nil {
			panic(err)
		}
		var driver driverListItem
		if len(networks) > 0 {
			driver_name := networks[0].(*models.NetworkConfig).Driver
			for index, driver_list_item := range driver_collection {
				if driver_name == driver_list_item.Name {
					driver = driver_collection[index]
				}
			}
		}
		newInterface <- interfaceItem{ConfigKey: config_key, Driver: driver}
	}
}

func init() {
	revel.OnAppStart(Init)
	//config, err := revel.Config.LoadConfig("app.conf")
	//fmt.Printf("config: %v %v\n", config, err)
	//fmt.Println("driver init!")
	//revel.Config.SetSection("dev")
	//result := revel.Config.Options("")
	//fmt.Println("net0 found: ",result)
	//fmt.Println("net0 result: ", result, " ", found)
	//revel.RegisterPlugin(network_manager{})
}
