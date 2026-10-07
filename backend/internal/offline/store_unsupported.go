//go:build !linux && !darwin

package offline

import (
	"errors"
	"os"
)

func lockStore(*os.Root) (*os.File, error) {
	return nil, errors.New("local durable store unsupported on this host")
}
func openSave(*os.Root, string) (*os.File, error) {
	return nil, errors.New("local durable store unsupported on this host")
}
