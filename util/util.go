package util

import (
	"fmt"
)

func Wrap(label string, err error) error {
	return fmt.Errorf("%s: %w", label, err)
}

func Deduplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]struct{})
	var list []T
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = struct{}{}
			list = append(list, item)
		}
	}
	return list
}
