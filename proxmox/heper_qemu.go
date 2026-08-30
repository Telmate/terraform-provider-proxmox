package proxmox

import (
	"net"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func parseCloudInitInterface(ipConfig pveSDK.CloudInitNetworkConfig, ciCustom, skipIPv4, skipIPv6 bool) (conn connectionInfo) {
	conn.SkipIPv4 = skipIPv4
	conn.SkipIPv6 = skipIPv6
	if ipConfig.IPv4 != nil {
		if ipConfig.IPv4.Address != nil {
			splitCIDR := strings.Split(string(*ipConfig.IPv4.Address), "/")
			conn.IPs.IPv4 = splitCIDR[0]
		}
	} else if !ciCustom {
		conn.SkipIPv4 = true
	}
	if ipConfig.IPv6 != nil {
		if ipConfig.IPv6.Address != nil {
			splitCIDR := strings.Split(string(*ipConfig.IPv6.Address), "/")
			conn.IPs.IPv6 = splitCIDR[0]
		}
	} else if !ciCustom {
		conn.SkipIPv6 = true
	}
	return
}

type primaryIPs struct {
	IPv4 string
	IPv6 string
}

type connectionInfo struct {
	IPs      primaryIPs
	SkipIPv4 bool
	SkipIPv6 bool
}

const (
	errorNoIPSummary   = "no IP config is found"
	errorNoIPv4Summary = "no IPv4 address is found"
	errorNoIPv6Summary = "no IPv6 address is found"
)

func (conn connectionInfo) agentDiagnostics(key, prefixSummary, prefixDetail string) diag.Diagnostics {
	if conn.IPs.IPv4 == "" {
		if conn.IPs.IPv6 == "" {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  prefixSummary + errorNoIPSummary,
				Detail:   prefixDetail + "no IP address was found before the time ran out, increasing '" + key + "' could resolve this issue."}}
		}
		if !conn.SkipIPv4 {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  prefixSummary + errorNoIPv4Summary,
				Detail:   prefixDetail + "no IPv4 address was found before the time ran out, increasing '" + key + "' could resolve this issue. To suppress this warning set '" + ip.RootSkipV4 + "' to true."}}
		}
		return diag.Diagnostics{}
	}
	if conn.IPs.IPv6 == "" && !conn.SkipIPv6 {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  prefixSummary + errorNoIPv6Summary,
			Detail:   prefixDetail + "no IPv6 address was found before the time ran out, increasing '" + key + "' could resolve this issue. To suppress this warning set '" + ip.RootSkipV6 + "' to true."}}
	}
	return diag.Diagnostics{}
}

func (conn connectionInfo) hasRequiredIP() bool {
	if conn.IPs.IPv4 == "" && !conn.SkipIPv4 || conn.IPs.IPv6 == "" && !conn.SkipIPv6 {
		return false
	}
	return true
}

func (conn *connectionInfo) parsePrimaryIPs(ipAddresses []net.IP) {
	for i := range ipAddresses {
		if ipAddresses[i].IsGlobalUnicast() {
			if ipAddresses[i].To4() != nil {
				if conn.IPs.IPv4 == "" {
					conn.IPs.IPv4 = ipAddresses[i].String()
				}
			} else {
				if conn.IPs.IPv6 == "" {
					conn.IPs.IPv6 = ipAddresses[i].String()
				}
			}
		}
	}
}
