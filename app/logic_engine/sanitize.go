package logic_engine

import (
	"strconv"
)

func floatify(in interface{}) float64 {
	var result float64
	var err error
	switch v := in.(type) {
	case string:
		result, err = strconv.ParseFloat(v, 64)
		if err != nil {
			LOG.Println(err)
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

func sanitize(obj *Object_t) {
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
			PTypes[k] == "timezone",
			PTypes[k] == "port":
			PValues = append(PValues, stringify(v))
		case PTypes[k] == "int":
			PValues = append(PValues, intify(v))
		}
	}
	(*obj)["PropertyValues"] = PValues
	(*obj)["Output"] = floatify((*obj)["Output"])
}
