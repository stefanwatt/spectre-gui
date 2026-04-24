package utils

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/log"
)

func ReflectToInt(iface interface{}) int {
	if iface == nil {
		log.Debug("could not reflect type:<nil> to int")
		return 0
	}

	switch v := iface.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	}

	val := reflect.ValueOf(iface)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(val.Uint())
	case reflect.Float32, reflect.Float64:
		return int(val.Float())
	}

	log.Debug(fmt.Sprintf("could not reflect type:%s to int", reflect.TypeOf(iface).String()))
	return 0
}

// ReflectToFloat converts interface{} to float64
func ReflectToFloat(iface interface{}) float64 {
	i, ok := iface.(float64)
	if ok {
		return i
	}
	u, ok := iface.(float32)
	if ok {
		return float64(u)
	}

	return float64(ReflectToInt(iface))
}

// IsZero determines if the value of interface{} is zero
func IsZero(d interface{}) bool {
	if d == nil {
		return false
	}
	switch a := d.(type) {
	case int64:
		if a == 0 {
			return true
		}
	case uint64:
		if a == 0 {
			return true
		}
	}
	return false
}

// IsTrue determines the truth of the value of interface{}
func IsTrue(d interface{}) bool {
	if d == nil {
		return false
	}
	switch a := d.(type) {
	case int64:
		if a == 1 {
			return true
		}
	case uint64:
		if a == 1 {
			return true
		}
	}
	return false
}
