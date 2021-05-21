package iprange

import (
	"context"
	"net"
)

type IPNet net.IPNet

var _ List = &IPNet{}
var _ Contiguous = &IPNet{}

// Broadcast is the ipv4 broadcast address for a network.
// 192.168.0.0/24 -> 192.168.0.255
func (ipn *IPNet) Broadcast() net.IP {
	if !isV4(ipn.IP) {
		return nil
	}
	return ipn.lastAddr()
}

func (ipn *IPNet) lastAddr() net.IP {
	mask := ipn.Mask
	network := ipn.IP.Mask(mask)

	var newMask net.IPMask

	newMask = make(net.IPMask, len(mask))
	for i := range newMask {
		newMask[i] = network[i] | ^mask[i]
	}

	bcast := net.IP(newMask)
	return bcast
}

// NextIPNet returns the next adjacent ip network
// 192.168.0.0/24 -> 192.168.1.0/24
func (ipn *IPNet) NextIPNet() *IPNet {
	if !isV4(ipn.IP) {
		return nil
	}
	ipA := ip2Bound(ipn.Broadcast())
	ipA.Inc()
	ret := &IPNet{IP: ipA.IP()}
	ret.Mask = append(ret.Mask, (*net.IPNet)(ipn).Mask...)
	return ret
}

// ContainsList returns true if the IPNet contains all the addresses in the List
func (ipn *IPNet) ContainsList(ctx context.Context, r List) (bool, error) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	switch r.(type) {
	case Contiguous:
		return ipn.ContainsContiguous(ctx, r.(Contiguous))
	}

	// TODO: fold out generic
	for ip := range r.IPs(childCtx) {
		ok := (*net.IPNet)(ipn).Contains(ip)
		if !ok {
			return false, nil
		}
	}
	return true, ctx.Err() // if the parent context was canceled
}

// IPs returns a channel of all the IP addresses in the network
func (ipn *IPNet) IPs(ctx context.Context) <-chan net.IP {
	out := ipChan()
	startA, endA := getBounds(ipn)
	go func() {
		defer close(out)
		for startA.LessOrEqualTo(endA) {
			select {
			case <-ctx.Done():
				return
			case out <- startA.IP():
			}
			startA.Inc()
		}
	}()
	return out
}

// ContainsList returns true if the IPNet contains all the addresses
func (ipn *IPNet) ContainsContiguous(ctx context.Context, b Contiguous) (bool, error) {
	return containsContiguous(ipn, b)
}

// Start returns the first IP in an IPNet. This will not be the first usable IP, but the network number.
func (ipn *IPNet) Start() net.IP { return ipn.IP }

// End returns the last IP in an IPNet. This will not be the last usable IP, but the broadcast.
func (ipn *IPNet) End() net.IP { return ipn.lastAddr() }
