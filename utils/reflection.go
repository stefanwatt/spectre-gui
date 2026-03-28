package utils

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/log"
)

func ReflectToInt(iface interface{}) int {
	i, ok := iface.(int64)
	if ok {
		return int(i)
	}
	j, ok := iface.(uint64)
	if ok {
		return int(j)
	}
	k, ok := iface.(int)
	if ok {
		return int(k)
	}
	l, ok := iface.(uint)
	if ok {
		return int(l)
	}
	b, ok := iface.(byte)
	if ok {
		return int(b)
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
