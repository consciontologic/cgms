package presence

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestLocalFallbackExpiresAndBounds(t *testing.T) {
	p, e := New("", Options{Capacity: 2, TTL: 30 * time.Second})
	if e != nil {
		t.Fatal(e)
	}
	if p.Touch(context.Background(), "a") {
		t.Fatal("no redis")
	}
	p.Touch(context.Background(), "b")
	p.Touch(context.Background(), "c")
	if p.Size() != 2 {
		t.Fatal(p.Size())
	}
	if online, source := p.Online(context.Background(), "a"); !online || source != "local" {
		t.Fatal(online, source)
	}
	p.now = func() time.Time { return time.Now().Add(time.Minute) }
	if online, _ := p.Online(context.Background(), "a"); online || p.Size() != 0 {
		t.Fatal("expiry")
	}
}
func TestUnreachableRedisFallsBack(t *testing.T) {
	p, e := New("redis://127.0.0.1:1", Options{Timeout: 10 * time.Millisecond})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if p.Touch(ctx, "a") {
		t.Fatal("unreachable success")
	}
	v, source := p.Online(ctx, "a")
	if !v || source != "local" {
		t.Fatal(v, source)
	}
}

func TestBlackholedRedisAndSaturationAreBounded(t *testing.T) {
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer listener.Close()
	accepted := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		c, e := listener.Accept()
		if e != nil {
			return
		}
		defer c.Close()
		close(accepted)
		<-release
	}()
	defer func() { close(release); <-finished }()
	p, e := New("redis://"+listener.Addr().String(), Options{Timeout: time.Second, MaxConcurrent: 1})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool, 1)
	go func() { done <- p.Touch(ctx, "a") }()
	<-accepted
	start := time.Now()
	if p.Touch(context.Background(), "b") {
		t.Fatal("saturated remote succeeded")
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatal("saturated call waited")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blackhole ignored cancellation")
	}
}
