package sim

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"io"
	"os"
	"syscall"
)

var ErrResourceBudget = errors.New("resource budget exhausted")

// checkRationalDigits runs before decoding arbitrary precision values. It never
// constructs a big integer and counts decimal digits, excluding a minus sign.
func checkRationalDigits(raw []byte, max int) error {
	return checkRationalDigitsLimit(raw, max, canonical.MaxBytes)
}
func checkRationalDigitsLimit(raw []byte, max, byteLimit int) error {
	if max < 1 {
		return fmt.Errorf("invalid rational digit budget")
	}
	var value any
	if err := canonical.DecodeLimit(raw, &value, byteLimit); err != nil {
		return err
	}
	var walk func(any) error
	walk = func(v any) error {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "numerator" || k == "denominator" {
					s, ok := v.(string)
					if !ok || s == "" {
						return errors.New("rational integer string required")
					}
					digits := s
					if digits[0] == '-' {
						if k == "denominator" {
							return errors.New("positive denominator required")
						}
						digits = digits[1:]
					}
					if len(digits) > max {
						return ErrResourceBudget
					}
					if digits == "" || len(digits) > 1 && digits[0] == '0' || s == "-0" || k == "denominator" && digits == "0" {
						return errors.New("noncanonical rational integer")
					}
					for _, c := range digits {
						if c < '0' || c > '9' {
							return errors.New("invalid rational integer")
						}
					}
				}
				if err := walk(v); err != nil {
					return err
				}
			}
		case []any:
			for _, v := range x {
				if err := walk(v); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(value)
}

// readArtifact accepts regular files only, avoiding uninterruptible FIFO/device
// reads. Cancellation is checked between bounded 32KiB reads and before return.
func readArtifact(ctx context.Context, path string, maxBytes int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if maxBytes < 1 {
		return nil, ErrResourceBudget
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("artifact must be a regular file")
	}
	if info.Size() > int64(maxBytes) {
		return nil, ErrResourceBudget
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("artifact must be a regular file")
	}
	out := make([]byte, 0)
	buf := make([]byte, 32768)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, e := f.Read(buf)
		if n > maxBytes-len(out) {
			return nil, ErrResourceBudget
		}
		out = append(out, buf[:n]...)
		if e != nil {
			if e != io.EOF {
				return nil, e
			}
			break
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// serializationBudget checks exact already-serialized payload sizes without
// overflow. Callers include all framing bytes and the prospective transition.
func serializationBudget(limit int, parts ...[]byte) error {
	if limit < 0 {
		return ErrResourceBudget
	}
	for _, p := range parts {
		if len(p) > limit {
			return ErrResourceBudget
		}
		limit -= len(p)
	}
	return nil
}
