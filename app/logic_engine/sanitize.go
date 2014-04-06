package logic_engine

import (
	"fmt"
	"strconv"
	"strings"
	"timecl/app/network_manager"
)

func floatify(in interface{}) float64 {
	var result float64
	var err error
	switch v := in.(type) {
	case string:
		result, err = strconv.ParseFloat(v, 64)
		if err != nil {
			PublishOneError(fmt.Errorf("Error parsing float from string:", err))
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
			PublishOneError(fmt.Errorf("Error parsing int from string:", err))
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
	(*obj)["Id"] = intify((*obj)["Id"])
	(*obj)["Source"] = intify((*obj)["Source"])

	var PCount int
	PCount = intify((*obj)["PropertyCount"])
	(*obj)["PropertyCount"] = PCount

	PNames := make([]interface{}, 0)
	for _, v := range (*obj)["PropertyNames"].([]interface{}) {
		PNames = append(PNames, stringify(v))
	}
	(*obj)["PropertyNames"] = PNames

	PTypes := make([]interface{}, 0)
	for _, v := range (*obj)["PropertyTypes"].([]interface{}) {
		PTypes = append(PTypes, stringify(v))
	}
	(*obj)["PropertyTypes"] = PTypes

	PValues := make([]interface{}, 0)
	for k, v := range (*obj)["PropertyValues"].([]interface{}) {
		switch {
		case PTypes[k] == "float":
			PValues = append(PValues, floatify(v))
		case PTypes[k] == "string",
			PTypes[k] == "time",
			PTypes[k] == "timezone":
			PValues = append(PValues, stringify(v))
		case PTypes[k] == "port":
			pval := toPortURI(v)
			var defaultURI network_manager.PortURI
			if pval == defaultURI {
				PValues = append(PValues, nil)
			} else {
				PValues = append(PValues, pval)
			}
		case PTypes[k] == "int":
			PValues = append(PValues, intify(v))
		}
	}
	(*obj)["PropertyValues"] = PValues
	(*obj)["Output"] = floatify((*obj)["Output"])
}
