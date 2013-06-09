package drivers

import (
	"fmt"
	//"github.com/robfig/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/robfig/revel/modules/jobs/app/jobs"
	"timecl/app/network_manager"
)

type DummyDriver struct {

}

func (p DummyDriver) Init(port string) {
    fmt.Println("Init Dummy Driver")
}
    
func init() {
	fmt.Println("Dummy Driver")
	network_manager.RegisterDriver("dummy", DummyDriver{})
}
