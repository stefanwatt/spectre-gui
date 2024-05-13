package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

func MapArrayConcurrent[T any, U any](items []T, transform func(T) U) []U {
	results := make([]U, len(items))
	var wg sync.WaitGroup
	wg.Add(len(items))

	// Result channel to collect transformation results in order
	type result struct {
		index int
		value U
	}
	resultChan := make(chan result, len(items))

	for i, item := range items {
		go func(index int, val T) {
			defer wg.Done()
			transformedValue := transform(val)
			resultChan <- result{index: index, value: transformedValue}
		}(i, item)
	}

	// Close the channel in a goroutine after all goroutines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	for res := range resultChan {
		results[res.index] = res.value
	}

	return results
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

func GetLastSubdirAndFilename(absolutePath string) string {
	lastSubdir := filepath.Base(filepath.Dir(absolutePath))
	return lastSubdir + "/" + filepath.Base(absolutePath)
}

func RetryCommand(command string, args []string, retries int, delay time.Duration) (*string, error) {
	Log(fmt.Sprintf("Attempting to run %s", command))
	var err error
	for i := 0; i < retries; i++ {
		cmd := exec.Command(command, args...)
		cmd.Stderr = os.Stderr
		bytes, cmdErr := cmd.Output()

		if cmdErr == nil {
			output := string(bytes)
			Log(fmt.Sprintf("Successfully ran %s", command))
			Log(fmt.Sprintf("Output: %s", output))
			return &output, nil
		}
		Log("Attempt %d failed: %s\n", i+1, cmdErr)
		Log("Command: %s %s\n", command, strings.Join(args, " "))

		if i < retries-1 {
			Log("Waiting before retry...")
			time.Sleep(delay)
		}

		err = cmdErr
	}

	return nil, fmt.Errorf("command failed after %d attempts with error: %s", retries, err)
}

func RandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err // Return the error if there's a problem generating the random number
		}
		result[i] = letters[num.Int64()]
	}
	return string(result), nil
}

func SliceString(s string, index int) string {
	runes := []rune(s)
	if index < 0 || index > len(runes) {
		return ""
	}
	return string(runes[index:])
}
