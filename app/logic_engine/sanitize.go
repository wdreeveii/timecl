package logic_engine

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/wdreeveii/timecl/app/logger"
	"github.com/wdreeveii/timecl/app/network_manager"
)

func floatify(in interface{}) float64 {
	var result float64
	var err error
	switch v := in.(type) {
	case string:
		result, err = strconv.ParseFloat(v, 64)
		if err != nil {
			logger.PublishOneError(fmt.Errorf("Error parsing float from string:", err))
		}
	case float64:
		result = v
	case int:
		result = float64(v)
	}
	return result
}

func intify(in interface{}) int {
	var result int
	switch v := in.(type) {
	case string:
		res, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			logger.PublishOneError(fmt.Errorf("Error parsing int from string:", err))
		}
		result = int(res)
	case float64:
		result = int(v)
	case int:
		result = v
	}
	return result
}

func stringify(in interface{}) string {
	var result string
	switch v := in.(type) {
	case float64:
		result = strconv.FormatFloat(v, 'f', 3, 64)
	case string:
		result = v
	case int:
		result = strconv.FormatInt(int64(v), 10)
	}
	return result
}
func toPortURI(in interface{}) (result network_manager.PortURI) {
	switch v := in.(type) {
	case string:
		split := strings.SplitN(v, "-", 5)
		if len(split) >= 4 {
			result = network_manager.PortURI{Network: intify(split[0]), Bus: intify(split[1]), Device: intify(split[2]), Port: intify(split[3])}
		}
	case network_manager.PortURI:
		result = v
	}
	return
}

func sanitize(obj *Object_t) {
	PValues := make([]interface{}, 0)
	for k, v := range obj.PropertyValues {
		switch {
		case obj.PropertyTypes[k] == "float":
			PValues = append(PValues, floatify(v))
		case obj.PropertyTypes[k] == "string",
			obj.PropertyTypes[k] == "time",
			obj.PropertyTypes[k] == "timezone":
			PValues = append(PValues, stringify(v))
		case obj.PropertyTypes[k] == "port":
			pval := toPortURI(v)
			var defaultURI network_manager.PortURI
			if pval == defaultURI {
				PValues = append(PValues, nil)
			} else {
				PValues = append(PValues, pval)
			}
		case obj.PropertyTypes[k] == "int":
			PValues = append(PValues, intify(v))
		}
	}
	obj.PropertyValues = PValues
}
