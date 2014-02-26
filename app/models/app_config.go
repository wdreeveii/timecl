
package models

import (
	"fmt"
	//"github.com/revel/revel"
	//"regexp"
)

type AppConfig struct {
	ConfigId	int
	Key			string
	Val			string
}

func (c AppConfig) String() string {
	return fmt.Sprintf("Config(%s)", c.Key)
}

