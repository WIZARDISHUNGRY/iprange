package iprange

import (
	"context"
	"fmt"
	"net"
	"testing"
)

func TestNextIPNet(t *testing.T) {
	ipnet := (*IPNet)(homeNetwork)
	next := ipnet.NextIPNet()
	if next == nil {
		t.Fatalf("no next ipnet")
	}
	if !next.IP.Equal(net.ParseIP("192.168.2.0")) ||
		next.Mask.String() != "ffffff00" {
		t.Fatalf("next ip %v", next)
	}
}

func TestIPNetIPS(t *testing.T) {
	ctx := context.Background()
	ipnet := (*IPNet)(homeNetwork)

	size, length := ipnet.Mask.Size()
	if size != 24 {
		t.Fatalf("size != 24 %v", size)
	}
	if length != 32 {
		t.Fatalf("length != 32 %v", length)
	}
	count := 0
	for ip := range ipnet.IPs(ctx) {
		count++
		if !(*net.IPNet)(ipnet).Contains(ip) {
			t.Fatalf("Contains %v", ip)
		}
	}
	if count != 256 {
		t.Fatalf("count != 256 %v", count)

	}
}

func TestIPV6(t *testing.T) {
	ctx := context.Background()
	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			if isV4(ipnet.IP) {
				continue
			}
			count := 0
			continue
			for ip := range (*IPNet)(ipnet).IPs(ctx) {
				if count%10000000 == 0 {
					fmt.Println(count, ip)
				}
				count++
			}
			fmt.Println(iface, ipnet.IP, count)

		}
	}
}
