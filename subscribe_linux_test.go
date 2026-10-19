//go:build linux
// +build linux

package netlink

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func openFDs(t *testing.T) int {
	t.Helper()
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	return len(fds)
}

// A subscription that ends on a receive error closes its socket, also
// while done is still open, so subscribing again does not leak it.
func TestSubscribeClosesSocketOnReceiveError(t *testing.T) {
	// A receive timeout makes Receive fail with EAGAIN when no update comes.
	timeout := &unix.Timeval{Usec: 1000}
	ignore := func(error) {}
	for _, c := range []struct {
		name string
		run  func(done <-chan struct{}) error
	}{
		{"link", func(done <-chan struct{}) error {
			ch := make(chan LinkUpdate)
			if err := LinkSubscribeWithOptions(ch, done, LinkSubscribeOptions{ErrorCallback: ignore, ReceiveTimeout: timeout}); err != nil {
				return err
			}
			for range ch {
			}
			return nil
		}},
		{"addr", func(done <-chan struct{}) error {
			ch := make(chan AddrUpdate)
			if err := AddrSubscribeWithOptions(ch, done, AddrSubscribeOptions{ErrorCallback: ignore, ReceiveTimeout: timeout}); err != nil {
				return err
			}
			for range ch {
			}
			return nil
		}},
		{"route", func(done <-chan struct{}) error {
			ch := make(chan RouteUpdate)
			if err := RouteSubscribeWithOptions(ch, done, RouteSubscribeOptions{ErrorCallback: ignore, ReceiveTimeout: timeout}); err != nil {
				return err
			}
			for range ch {
			}
			return nil
		}},
		{"neigh", func(done <-chan struct{}) error {
			ch := make(chan NeighUpdate)
			if err := NeighSubscribeWithOptions(ch, done, NeighSubscribeOptions{ErrorCallback: ignore, ReceiveTimeout: timeout}); err != nil {
				return err
			}
			for range ch {
			}
			return nil
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			done := make(chan struct{})
			defer close(done)
			before := openFDs(t)
			const n = 20
			for i := 0; i < n; i++ {
				if err := c.run(done); err != nil {
					t.Fatal(err)
				}
			}
			// The socket is closed by a goroutine: wait for it.
			deadline := time.Now().Add(5 * time.Second)
			for openFDs(t) > before && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			if got := openFDs(t); got > before {
				t.Errorf("%d subscriptions ended, %d file descriptors still open", n, got-before)
			}
		})
	}
}
