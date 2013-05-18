
package models

import (
	"fmt"
	"github.com/robfig/revel"
	"regexp"
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

