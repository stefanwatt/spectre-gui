package main

import (
	"fmt"
)

func Find[T any](array []T, find_func func(T) bool) (T, error) {
	var zero T
	for _, value := range array {
		if find_func(value) {
			return value, nil
		}
	}
	return zero, fmt.Errorf("no match found")
}

func MapArray[T any, U any](array []T, map_func func(T) U) []U {
	var result []U
	for _, value := range array {
		result = append(result, map_func(value))
	}
	return result
}

func Filter[T any](array []T, filter_func func(T) bool) []T {
	var result []T
	for _, value := range array {
		if filter_func(value) {
			result = append(result, value)
		}
	}
	return result
}

func Flatten[T any](slice [][]T) []T {
	var flatSlice []T
	for _, innerSlice := range slice {
		flatSlice = append(flatSlice, innerSlice...)
	}
	return flatSlice
}
