// Package presence contains disposable hints only. No command, membership,
// allowance or recovery decision may depend on this cache.
package presence

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Capacity, MaxConcurrent int
	TTL, Timeout            time.Duration
}
type Tracker struct {
	mu                      sync.Mutex
	local                   map[string]time.Time
	options                 Options
	address, user, password string
	slots                   chan struct{}
	now                     func() time.Time
}

func New(raw string, o Options) (*Tracker, error) {
	if o.Capacity <= 0 {
		o.Capacity = 1024
	}
	if o.MaxConcurrent <= 0 {
		o.MaxConcurrent = 4
	}
	if o.TTL <= 0 {
		o.TTL = 30 * time.Second
	}
	if o.Timeout <= 0 {
		o.Timeout = 100 * time.Millisecond
	}
	if o.Capacity > 10000 || o.MaxConcurrent > 32 || o.TTL > time.Minute || o.Timeout > time.Second {
		return nil, errors.New("presence bounds exceeded")
	}
	p := &Tracker{local: map[string]time.Time{}, options: o, slots: make(chan struct{}, o.MaxConcurrent), now: time.Now}
	if raw != "" {
		u, e := url.Parse(raw)
		if e != nil || u.Scheme != "redis" || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/0") {
			return nil, errors.New("invalid Redis hint configuration")
		}
		port := u.Port()
		if port == "" {
			port = "6379"
		}
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return nil, errors.New("invalid Redis port")
		}
		p.address = net.JoinHostPort(u.Hostname(), port)
		if u.User != nil {
			p.user = u.User.Username()
			p.password, _ = u.User.Password()
		}
	}
	return p, nil
}
func key(raw string) (string, bool) {
	if raw == "" || len(raw) > 512 {
		return "", false
	}
	h := sha256.Sum256([]byte(raw))
	return "cgms:presence:" + hex.EncodeToString(h[:]), true
}
func (p *Tracker) prune(now time.Time) {
	for k, end := range p.local {
		if !now.Before(end) {
			delete(p.local, k)
		}
	}
}
func (p *Tracker) Size() int { p.mu.Lock(); defer p.mu.Unlock(); p.prune(p.now()); return len(p.local) }

// Touch retains bounded local presence even when Redis is unavailable. False
// means the shared hint was not updated, not that the account is disconnected.
func (p *Tracker) Touch(ctx context.Context, raw string) bool {
	k, ok := key(raw)
	if !ok {
		return false
	}
	p.mu.Lock()
	p.prune(p.now())
	if _, exists := p.local[k]; exists || len(p.local) < p.options.Capacity {
		p.local[k] = p.now().Add(p.options.TTL)
	}
	p.mu.Unlock()
	_, e := p.request(ctx, "SET", k, "1", "PX", strconv.FormatInt(max(1, p.options.TTL.Milliseconds()), 10))
	return e == nil
}
func (p *Tracker) Online(ctx context.Context, raw string) (bool, string) {
	k, ok := key(raw)
	if !ok {
		return false, "unknown"
	}
	v, e := p.request(ctx, "GET", k)
	if e == nil {
		return v == "1", "redis"
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune(p.now())
	_, exists := p.local[k]
	if exists {
		return true, "local"
	}
	return false, "unknown"
}
func (p *Tracker) request(ctx context.Context, args ...string) (string, error) {
	if p.address == "" {
		return "", errors.New("no shared presence")
	}
	select {
	case p.slots <- struct{}{}:
		defer func() { <-p.slots }()
	default:
		return "", errors.New("presence saturated")
	}
	ctx, cancel := context.WithTimeout(ctx, p.options.Timeout)
	defer cancel()
	conn, e := (&net.Dialer{}).DialContext(ctx, "tcp", p.address)
	if e != nil {
		return "", errors.New("presence unavailable")
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if e = conn.SetDeadline(deadline); e != nil {
		return "", e
	}
	reader := bufio.NewReader(io.LimitReader(conn, 4096))
	call := func(words []string) (string, error) {
		var b strings.Builder
		fmt.Fprintf(&b, "*%d\r\n", len(words))
		for _, w := range words {
			fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(w), w)
		}
		if _, e = io.WriteString(conn, b.String()); e != nil {
			return "", errors.New("presence unavailable")
		}
		line, e := reader.ReadString('\n')
		if e != nil || !strings.HasSuffix(line, "\r\n") {
			return "", errors.New("invalid presence response")
		}
		switch line[0] {
		case '+':
			return strings.TrimSuffix(line[1:], "\r\n"), nil
		case '$':
			n, e := strconv.Atoi(strings.TrimSuffix(line[1:], "\r\n"))
			if e != nil || n < -1 || n > 1024 {
				return "", errors.New("invalid presence response")
			}
			if n == -1 {
				return "", nil
			}
			v := make([]byte, n+2)
			if _, e = io.ReadFull(reader, v); e != nil || string(v[n:]) != "\r\n" {
				return "", errors.New("invalid presence response")
			}
			return string(v[:n]), nil
		default:
			return "", errors.New("presence unavailable")
		}
	}
	if p.password != "" {
		auth := []string{"AUTH", p.password}
		if p.user != "" {
			auth = []string{"AUTH", p.user, p.password}
		}
		if _, e = call(auth); e != nil {
			return "", e
		}
	}
	return call(args)
}
