// Package config validates dedicated service settings before any connections.
package config

import (
	"encoding/json"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Observability struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
}
type Config struct {
	EconomyEnabled  bool          `json:"economy_enabled,omitempty"`
	MaxStreams      int           `json:"max_streams,omitempty"`
	ProxyClientIP   bool          `json:"proxy_client_ip,omitempty"`
	Version         int           `json:"version"`
	Listen          string        `json:"listen"`
	Origin          string        `json:"origin"`
	InsecureLocal   bool          `json:"insecure_local"`
	RulesFile       string        `json:"rules_file"`
	DatabaseURLFile string        `json:"database_url_file"`
	RedisURLFile    string        `json:"redis_url_file,omitempty"`
	Observability   Observability `json:"observability"`
}
type Secrets struct{ DatabaseURL, RedisURL string }

func Decode(r io.Reader) (Config, error) {
	var c Config
	raw, err := io.ReadAll(io.LimitReader(r, 65537))
	if err != nil || canonical.DecodeLimit(raw, &c, 65536) != nil {
		return c, errors.New("invalid service configuration")
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	for _, name := range []string{"proxy_client_ip", "redis_url_file", "economy_enabled"} {
		if value, present := fields[name]; present && string(value) == "null" {
			return c, errors.New("invalid optional service setting")
		}
	}
	for _, name := range []string{"version", "listen", "origin", "insecure_local", "rules_file", "database_url_file", "observability"} {
		if value, ok := fields[name]; !ok || string(value) == "null" {
			return c, errors.New("missing service setting")
		}
	}
	var observation map[string]json.RawMessage
	_ = json.Unmarshal(fields["observability"], &observation)
	for _, name := range []string{"enabled", "endpoint"} {
		if value, ok := observation[name]; !ok || string(value) == "null" {
			return c, errors.New("missing observability setting")
		}
	}
	if value, present := fields["max_streams"]; present {
		if string(value) == "null" || c.MaxStreams < 1 || c.MaxStreams > 4096 {
			return c, errors.New("invalid stream capacity")
		}
	} else {
		c.MaxStreams = 64
	}
	u, e := url.Parse(c.Origin)
	_, port, listenErr := net.SplitHostPort(c.Listen)
	number, portErr := strconv.Atoi(port)
	if portErr != nil || number < 1 || number > 65535 {
		return c, errors.New("invalid listener port")
	}
	if c.Version != 1 || e != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(c.InsecureLocal && u.Scheme == "http")) || listenErr != nil || port == "" || c.RulesFile == "" || c.DatabaseURLFile == "" {
		return c, errors.New("invalid service settings")
	}
	if c.Observability.Enabled {
		u, e = url.Parse(c.Observability.Endpoint)
		if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
			return c, errors.New("invalid telemetry endpoint")
		}
	}
	return c, nil
}
func Load(path string) (Config, error) {
	f, e := os.Open(path)
	if e != nil {
		return Config{}, errors.New("service configuration unavailable")
	}
	defer f.Close()
	return Decode(f)
}
func readSecret(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", errors.New("secret unavailable")
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0007 != 0 {
		return "", errors.New("secret must be a private regular file")
	}
	raw, e := io.ReadAll(io.LimitReader(f, 8193))
	if e != nil || len(raw) > 8192 || len(strings.TrimSpace(string(raw))) == 0 {
		return "", errors.New("invalid secret")
	}
	return strings.TrimSpace(string(raw)), nil
}
func (c Config) Secrets() (Secrets, error) {
	var s Secrets
	var e error
	s.DatabaseURL, e = readSecret(c.DatabaseURLFile)
	if e != nil {
		return s, e
	}
	if c.RedisURLFile != "" {
		s.RedisURL, e = readSecret(c.RedisURLFile)
	}
	return s, e
}
