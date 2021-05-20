package iprange

import (
	"context"
	"net"
	"testing"
)

var homeNetworkNumber, homeNetwork, _ = net.ParseCIDR("192.168.1.0/24")
var tinyNetNumber, tinyNetwork, _ = net.ParseCIDR("192.168.1.20/26")

func TestContainsList(t *testing.T) {
	testCases := []struct {
		desc           string
		Range1, Range2 func() List
		expect         bool
	}{
		{
			desc:   "base",
			Range1: func() List { return (*IPNet)(homeNetwork) },
			Range2: func() List { return (*IPNet)(tinyNetwork) },
			expect: true,
		},
		{
			desc:   "range",
			Range1: func() List { return (*IPNet)(homeNetwork) },
			Range2: func() List { return FromIPRange(net.ParseIP("192.168.1.1"), net.ParseIP("192.168.1.100")) },

			expect: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.Background()
			r1 := tC.Range1()
			r2 := tC.Range2()
			ok, err := r1.ContainsList(ctx, r2)
			if err != nil {
				t.Fatalf("ContainsList error %v", err)
			}
			if ok != tC.expect {
				t.Fatalf("ContainsList ok != expect %v != %v", ok, tC.expect)
			}
		})
	}
}

func BenchmarkContainsList_tiny(b *testing.B) {
	r1 := FromIPRange(net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255"))
	r2 := (*IPNet)(tinyNetwork)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := r1.ContainsList(ctx, r2)
		if err != nil {
			b.Fatalf("ContainsList error %v", err)
		}
	}
}

func BenchmarkContainsList_Everything(b *testing.B) {
	r1 := FromIPRange(net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255"))
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := r1.ContainsList(ctx, r1)
		if err != nil {
			b.Fatalf("ContainsList error %v", err)
		}
	}
}
