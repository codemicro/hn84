package util

import (
	"fmt"
)

func Wrap(label string, err error) error {
	return fmt.Errorf("%s: %w", label, err)
}
