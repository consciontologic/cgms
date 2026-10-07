// offlinehost is a host-only stdin/stdout harness. It opens no sockets and is
// deliberately not the Android/iOS packaging or Flutter web offline runtime.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/offline"
	"io"
	"os"
)

const maxRequestBytes = 64 << 10

func serve(input io.Reader, output io.Writer, host *offline.Host) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), maxRequestBytes+1)
	writer := json.NewEncoder(output)
	for scanner.Scan() {
		var request offline.Request
		if err := canonical.DecodeLimit(scanner.Bytes(), &request, maxRequestBytes); err != nil {
			if err = writer.Encode(map[string]string{"error": "invalid_request"}); err != nil {
				return err
			}
			continue
		}
		view, err := host.Handle(request)
		if err != nil {
			code := "local_request_rejected"
			if errors.Is(err, offline.ErrDurabilityUnknown) {
				code = "save_outcome_unknown_reload"
			}
			if err = writer.Encode(map[string]string{"error": code}); err != nil {
				return err
			}
			continue
		}
		if err = writer.Encode(map[string]any{"schema": "cgms-offline-view-v1", "view": view}); err != nil {
			return err
		}
	}
	return scanner.Err()
}
func run() error {
	directory := flag.String("save-dir", "", "existing private app save directory")
	rules := flag.String("rules-hash", "", "build-provided canonical rules SHA-256")
	flag.Parse()
	if *directory == "" || len(*rules) != 64 {
		return errors.New("save directory and rules SHA-256 required")
	}
	for _, c := range *rules {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return errors.New("invalid rules SHA-256")
		}
	}
	store, err := offline.OpenStore(*directory, *rules)
	if err != nil {
		return err
	}
	defer store.Close()
	return serve(os.Stdin, os.Stdout, offline.NewHost(store))
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "offline host stopped; reopen or check local configuration")
		os.Exit(1)
	}
}
