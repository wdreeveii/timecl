package drivers

import (
	"fmt"
	//"github.com/revel/revel"
	//"github.com/mewkiz/pkg/hashutil/crc16"
	//"github.com/revel/revel/modules/jobs/app/jobs"
	"timecl/app/network_manager"
)

type DummyDriver struct {
}

func (p *DummyDriver) Init(port string, network_id int) {
	fmt.Println("Init Dummy Driver")
}

func (p *DummyDriver) Stop() {
	fmt.Println("DUMMY STOP")
}

func (p *DummyDriver) Copy() network_manager.DriverInterface {
	a := new(DummyDriver)
	return a
}

func (p *DummyDriver) ListPorts() []network_manager.BusDef {
	return make([]network_manager.BusDef, 0)
}

func init() {
	fmt.Println("Dummy Driver")
	network_manager.RegisterDriver("dummy", new(DummyDriver))
}
