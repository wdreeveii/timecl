
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
}

type EmptyDriver struct {}

func (p EmptyDriver) Init(port string)		{}

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

var restartInterface = make(chan restartConfig)
var newInterface = make(chan interfaceItem)

//get (network number, bus number, device number, port number
func Get(NetworkID int, bus int, device int, port int) {
	
}

func Set(NetworkID int, bus int, device int, port int, value interface{}) {
	
}

func RestartDriver(NetworkID int, driver string) {
	restartInterface <- restartConfig{NetworkID: NetworkID, Driver: driver}
}

func interfacesManager() {
	var interfaces []interfaceItem
	for {
		var reConfig 		restartConfig
		var newIntConfig 	interfaceItem
		
		select {
		case reConfig = <- restartInterface:
			for _, val := range driver_collection {
				if val.Name == reConfig.Driver {
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
					fmt.Println("Starting driver....")
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
