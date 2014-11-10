package drivers

import (
	"github.com/wdreeveii/timecl/app/network_manager"
)

type DummyDriver struct {
}

func (p DummyDriver) Init(port string, network_id int) {
}

func (p DummyDriver) Stop() {
}

func (p DummyDriver) ListPorts() []network_manager.BusDef {
	return make([]network_manager.BusDef, 0)
}

func init() {
	network_manager.RegisterDriver("dummy", DummyDriver{})
}
