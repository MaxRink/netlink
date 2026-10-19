package netlink

import "github.com/vishvananda/netlink/nl"

// Family type definitions
const (
	FAMILY_ALL  = nl.FAMILY_ALL
	FAMILY_V4   = nl.FAMILY_V4
	FAMILY_V6   = nl.FAMILY_V6
	FAMILY_MPLS = nl.FAMILY_MPLS
)

// ErrDumpInterrupted is an alias for [nl.ErrDumpInterrupted].
var ErrDumpInterrupted = nl.ErrDumpInterrupted

// closeOnDone closes the subscription socket s when done is closed, or
// when the returned stop function is called, whichever comes first. The
// receive goroutine of a subscription calls stop when it exits, so a
// subscription that ends on a receive error does not keep its socket and
// this goroutine until done is closed (or forever with a nil done).
func closeOnDone(s *nl.NetlinkSocket, done <-chan struct{}) (stop func()) {
	exited := make(chan struct{})
	go func() {
		select {
		case <-done:
		case <-exited:
		}
		s.Close()
	}()
	return func() { close(exited) }
}
