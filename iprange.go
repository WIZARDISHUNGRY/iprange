package iprange

import (
	"context"
	"encoding/binary"
	"net"

	"github.com/shabbyrobe/go-num"
)

type List interface {
	ContainsList(context.Context, List) (bool, error)
	IPs(context.Context) <-chan net.IP
}
type Contiguous interface {
	List
	ContainsContiguous(context.Context, Contiguous) (bool, error)
	Start() net.IP
	End() net.IP
}

const ipChanSize = 64

func ipChan() chan net.IP { return make(chan net.IP, ipChanSize) }

func ip2int(ip net.IP) ip4bound {
	if len(ip) == 16 {
		return ip4bound(binary.BigEndian.Uint32(ip[12:16]))
	}
	return ip4bound(binary.BigEndian.Uint32(ip))
}

func int2ip(nn ip4bound) net.IP {
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, uint32(nn))
	return ip
}

const halfV6 = net.IPv6len / 2

func ip2int_v6(ip net.IP) ip6bound {
	hi := binary.BigEndian.Uint64(ip)
	lo := binary.BigEndian.Uint64(ip[halfV6:])
	return ip6bound(num.U128FromRaw(hi, lo))
}

func int2ip_v6(nn ip6bound) net.IP {
	ip := make(net.IP, net.IPv6len)
	hi, lo := num.U128(nn).Raw()
	binary.BigEndian.PutUint64(ip, hi)
	binary.BigEndian.PutUint64(ip[halfV6:], lo)
	return ip
}

func ips(ctx context.Context, c Contiguous) <-chan net.IP {
	start, end := c.Start(), c.End()
	out := make(chan net.IP)
	startA, endA := ip2int(start), ip2int(end)
	go func() {
		defer close(out)
		for i := startA; i <= endA; i++ {
			select {
			case <-ctx.Done():
				return
			case out <- int2ip(i):
			}
		}
	}()
	return out
}

func containsContiguous(a, b Contiguous) (bool, error) {
	startA, endA := getBounds(a)
	startB, endB := getBounds(b)
	return startA.LessOrEqualTo(startB) &&
			endA.GreaterOrEqualTo(endB),
		nil
}

func getBounds(c Contiguous) (bound, bound) {
	return ip2Bound(c.Start()),
		ip2Bound(c.End())
}

func isV4(ip net.IP) bool {
	return ip.To4() != nil
}

type bound interface {
	Inc()
	LessOrEqualTo(b bound) bool
	GreaterOrEqualTo(b bound) bool
	IP() net.IP
}

func ip2Bound(ip net.IP) bound {
	if isV4(ip) {
		b := ip2int(ip)
		return &b
	}
	b := ip2int_v6(ip)
	return &b
}

type ip6bound num.U128

func (ip *ip6bound) Inc() {
	(*ip) = ip6bound((*num.U128)(ip).Add64(1))
}
func (ip *ip6bound) LessOrEqualTo(b bound) bool {
	return (*num.U128)(ip).LessOrEqualTo(
		*(*num.U128)(b.(*ip6bound)),
	)
}
func (ip *ip6bound) GreaterOrEqualTo(b bound) bool {
	return (*num.U128)(ip).GreaterOrEqualTo(
		*(*num.U128)(b.(*ip6bound)),
	)
}
func (ip *ip6bound) IP() net.IP {
	return int2ip_v6(*ip)
}

type ip4bound uint32

func (ip *ip4bound) Inc() {
	*(*uint32)(ip) += 1
}
func (ip *ip4bound) LessOrEqualTo(b bound) bool {
	return *(*uint32)(ip) <= *(*uint32)(b.(*ip4bound))
}
func (ip *ip4bound) GreaterOrEqualTo(b bound) bool {
	return *(*uint32)(ip) >= *(*uint32)(b.(*ip4bound))
}
func (ip *ip4bound) IP() net.IP {
	return int2ip(*ip)
}
