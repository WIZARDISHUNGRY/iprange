package iprange

import (
	"context"
	"net"
)

// Range represents a continuous range of IP addresses.
type Range struct {
	start, end net.IP
}

var _ List = &Range{}
var _ Contiguous = &Range{}

// FromIPRange creates a Range from a pair of IP addresses.
func FromIPRange(start, end net.IP) *Range {
	// TODO: validate addr family and ordering
	return &Range{start: start, end: end}
}

// ContainsList returns true if every list item is contained in the Range.
func (r *Range) ContainsList(ctx context.Context, list List) (bool, error) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	switch list.(type) {
	case Contiguous:
		return r.ContainsContiguous(ctx, list.(Contiguous))
	}

	startA, endA := getBounds(r)
	ok := true
	for ip := range list.IPs(childCtx) {
		ipA := ip2Bound(ip)
		if ipA.LessOrEqualTo(startA) || ipA.GreaterOrEqualTo(endA) {
			ok = false
			break
		}
	}
	return ok, ctx.Err() // if the parent context was canceled
}

// IPs returns a channel of every IP in the Range.
func (r *Range) IPs(ctx context.Context) <-chan net.IP {
	return ips(ctx, r)
}

// ContainsContiguous returns true if every list item is contained in the Range.
func (r *Range) ContainsContiguous(ctx context.Context, b Contiguous) (bool, error) {
	return containsContiguous(r, b)
}

// Start returns the first IP in the Range.
func (r *Range) Start() net.IP { return r.start }

// Last returns the last IP in the Range.
func (r *Range) End() net.IP { return r.end }
