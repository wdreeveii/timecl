
package models

import (
	"fmt"
	//"github.com/robfig/revel"
	//"regexp"
)

type Device struct {
	DeviceId	int
	Name		string
	Mac			int
	Address		int
	
}

func (d *Device) String() string {
	return fmt.Sprintf("Device(%s)", d.Name)
}

type NetworkConfig struct {
	NetworkID		int
	ConfigKey		string
	DevicePath		string
	Driver			string
}

func (n *NetworkConfig) String() string {
	return fmt.Sprintf("Network Config (%d, %s, %s, %s)", n.NetworkID, n.ConfigKey, n.DevicePath, n.Driver)
}
