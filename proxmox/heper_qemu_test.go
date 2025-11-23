package proxmox

import (
	"net"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
)

func Test_connectionInfo_agentDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		input  connectionInfo
		output string
	}{
		{name: `empty empty false false`,
			input:  connectionInfo{primaryIPs{"", ""}, false, false},
			output: errorGuestAgentNoIPSummary},
		{name: `empty empty false true`,
			input:  connectionInfo{primaryIPs{"", ""}, false, true},
			output: errorGuestAgentNoIPSummary},
		{name: `empty empty true false`,
			input:  connectionInfo{primaryIPs{"", ""}, true, false},
			output: errorGuestAgentNoIPSummary},
		{name: `empty empty true true`,
			input:  connectionInfo{primaryIPs{"", ""}, true, true},
			output: errorGuestAgentNoIPSummary},
		{name: `empty set false false`,
			input:  connectionInfo{primaryIPs{"", "set"}, false, false},
			output: errorGuestAgentNoIPv4Summary},
		{name: `empty set false true`,
			input:  connectionInfo{primaryIPs{"", "set"}, false, true},
			output: errorGuestAgentNoIPv4Summary},
		{name: `empty set true false`,
			input:  connectionInfo{primaryIPs{"", "set"}, true, false},
			output: ""},
		{name: `empty set true true`,
			input:  connectionInfo{primaryIPs{"", "set"}, true, true},
			output: ""},
		{name: `set empty false false`,
			input:  connectionInfo{primaryIPs{"set", ""}, false, false},
			output: errorGuestAgentNoIPv6Summary},
		{name: `set empty false true`,
			input:  connectionInfo{primaryIPs{"set", ""}, false, true},
			output: ""},
		{name: `set empty true false`,
			input:  connectionInfo{primaryIPs{"set", ""}, true, false},
			output: errorGuestAgentNoIPv6Summary},
		{name: `set empty true true`,
			input:  connectionInfo{primaryIPs{"set", ""}, true, true},
			output: ""},
		{name: `set set false false`,
			input:  connectionInfo{primaryIPs{"set", "set"}, false, false},
			output: ""},
		{name: `set set false true`,
			input:  connectionInfo{primaryIPs{"set", "set"}, false, true},
			output: ""},
		{name: `set set true false`,
			input:  connectionInfo{primaryIPs{"set", "set"}, true, false},
			output: ""},
		{name: `set set true true`,
			input:  connectionInfo{primaryIPs{"set", "set"}, true, true},
			output: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.output != "" {
				tmpDiag := test.input.agentDiagnostics()
				require.Len(t, tmpDiag, 1)
				require.Equal(t, test.output, tmpDiag[0].Summary)
			} else {
				require.Equal(t, diag.Diagnostics{}, test.input.agentDiagnostics())
			}
		})
	}
}

func Test_HasRequiredIP(t *testing.T) {
	tests := []struct {
		name   string
		input  connectionInfo
		output bool
	}{
		{name: `IPv4`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"}},
			output: false},
		{name: `IPv4 SkipIPv4`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"},
				SkipIPv4: true},
			output: false},
		{name: `IPv4 SkipIPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"},
				SkipIPv6: true},
			output: true},
		{name: `SkipIPv4`,
			input:  connectionInfo{},
			output: false},
		{name: `IPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}},
			output: false},
		{name: `IPv6 SkipIPv4`,
			input: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv4: true},
			output: true},
		{name: `IPv6 SkipIPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv6: true},
			output: false},
		{name: `SkipIPv6`,
			input:  connectionInfo{},
			output: false},
		{name: `IPv4 IPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}},
			output: true},
		{name: `IPv4 IPv6 SkipIPv4`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv4: true},
			output: true},
		{name: `IPv4 IPv6 SkipIPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv6: true},
			output: true},
		{name: `IPv4 IPv6 SkipIPv4 SkipIPv6`,
			input: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv4: true,
				SkipIPv6: true},
			output: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.hasRequiredIP())
		})
	}
}

func Test_ParseCloudInitInterface(t *testing.T) {
	type testInput struct {
		ci       pveSDK.CloudInitNetworkConfig
		ciCustom bool
		skipIPv4 bool
		skipIPv6 bool
	}
	tests := []struct {
		name   string
		input  testInput
		output connectionInfo
	}{
		{name: `IPv4=DHCP`,
			input: testInput{ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
				DHCP: true}}},
			output: connectionInfo{
				SkipIPv6: true}},
		{name: `IPv4=DHCP ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					DHCP: true}},
				ciCustom: true}},
		{name: `IPv4=DHCP SkipIPv4`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					DHCP: true}},
				skipIPv4: true},
			output: connectionInfo{
				SkipIPv4: true,
				SkipIPv6: true}},
		{name: `IPv4=DHCP SkipIPv4 ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					DHCP: true}},
				ciCustom: true,
				skipIPv4: true},
			output: connectionInfo{SkipIPv4: true}},
		{name: `IPv4=Static`,
			input: testInput{ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
				Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))}}},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"},
				SkipIPv6: true}},
		{name: `IPv4=Static ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))}},
				ciCustom: true},
			output: connectionInfo{IPs: primaryIPs{IPv4: "192.168.1.1"}}},
		{name: `IPv4=Static IPv6=Static`,
			input: testInput{ci: pveSDK.CloudInitNetworkConfig{
				IPv4: &pveSDK.CloudInitIPv4Config{
					Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))},
				IPv6: &pveSDK.CloudInitIPv6Config{
					Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}}},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}}},
		{name: `IPv4=Static IPv6=Static ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{
					IPv4: &pveSDK.CloudInitIPv4Config{
						Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))},
					IPv6: &pveSDK.CloudInitIPv6Config{
						Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}},
				ciCustom: true},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1",
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}}},
		{name: `IPv4=Static SkipIPv4`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))}},
				skipIPv4: true},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"},
				SkipIPv4: true,
				SkipIPv6: true}},
		{name: `IPv4=Static SkipIPv4 ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv4: &pveSDK.CloudInitIPv4Config{
					Address: util.Pointer(pveSDK.IPv4CIDR("192.168.1.1/24"))}},
				ciCustom: true,
				skipIPv4: true},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: "192.168.1.1"},
				SkipIPv4: true}},
		{name: `IPv6=DHCP`,
			input: testInput{ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
				DHCP: true}}},
			output: connectionInfo{SkipIPv4: true}},
		{name: `IPv6=DHCP ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					DHCP: true}},
				ciCustom: true}},
		{name: `IPv6=DHCP SkipIPv6`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					DHCP: true}},
				skipIPv6: true},
			output: connectionInfo{
				SkipIPv4: true,
				SkipIPv6: true}},
		{name: `IPv6=DHCP SkipIPv6 ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					DHCP: true}},
				ciCustom: true,
				skipIPv6: true},
			output: connectionInfo{SkipIPv6: true}},
		{name: `IPv6=Static`,
			input: testInput{ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
				Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}}},
			output: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv4: true}},
		{name: `IPv6=Static ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}},
				ciCustom: true},
			output: connectionInfo{IPs: primaryIPs{IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}}},
		{name: `IPv6=Static SkipIPv6`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}},
				skipIPv6: true},
			output: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv4: true,
				SkipIPv6: true}},
		{name: `IPv6=Static SkipIPv6 ciCustom`,
			input: testInput{
				ci: pveSDK.CloudInitNetworkConfig{IPv6: &pveSDK.CloudInitIPv6Config{
					Address: util.Pointer(pveSDK.IPv6CIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"))}},
				ciCustom: true,
				skipIPv6: true},
			output: connectionInfo{IPs: primaryIPs{
				IPv6: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				SkipIPv6: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, parseCloudInitInterface(test.input.ci, test.input.ciCustom, test.input.skipIPv4, test.input.skipIPv6))
		})
	}
}

func Test_ParsePrimaryIPs(t *testing.T) {
	parseIP := func(ip string) net.IP {
		realIP, _, _ := net.ParseCIDR(ip)
		return realIP
	}
	formatIP := func(ip string) string {
		return net.ParseIP(ip).String()
	}
	type testInput struct {
		nets []net.IP
		conn connectionInfo
	}
	tests := []struct {
		name   string
		input  testInput
		output connectionInfo
	}{
		{name: `Only IPv4`,
			input: testInput{
				nets: []net.IP{
					parseIP("127.0.0.1/8"),
					parseIP("192.168.1.1/24"),
					parseIP("::1/128")}},
			output: connectionInfo{IPs: primaryIPs{IPv4: formatIP("192.168.1.1")}}},
		{name: `Only IPv6`,
			input: testInput{
				nets: []net.IP{
					parseIP("127.0.0.1/8"),
					parseIP("::1/128"),
					parseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64")}},
			output: connectionInfo{IPs: primaryIPs{IPv6: formatIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")}}},
		{name: `Full test`,
			input: testInput{
				nets: []net.IP{
					parseIP("10.10.10.1/16"),
					parseIP("3ffe:1900:4545:3:200:f8ff:fe21:67cf/64")}},
			output: connectionInfo{IPs: primaryIPs{
				IPv4: formatIP("10.10.10.1"),
				IPv6: formatIP("3ffe:1900:4545:3:200:f8ff:fe21:67cf")}}},
		{name: `IPv4 Already Set`,
			input: testInput{
				nets: []net.IP{parseIP("192.168.1.1/24")},
				conn: connectionInfo{IPs: primaryIPs{IPv4: formatIP("10.10.1.1")}}},
			output: connectionInfo{IPs: primaryIPs{IPv4: formatIP("10.10.1.1")}}},
		{name: `IPv6 Already Set`,
			input: testInput{
				nets: []net.IP{parseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64")},
				conn: connectionInfo{IPs: primaryIPs{IPv6: formatIP("3ffe:1900:4545:3:200:f8ff:fe21:67cf")}}},
			output: connectionInfo{IPs: primaryIPs{IPv6: formatIP("3ffe:1900:4545:3:200:f8ff:fe21:67cf")}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.conn.parsePrimaryIPs(test.input.nets))
		})
	}
}
