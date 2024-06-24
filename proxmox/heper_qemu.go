package proxmox

import (
	"strings"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	ErrorGuestAgentNotRunning    string = "500 QEMU guest agent is not running"
	errorGuestAgentNoIPSummary   string = "Qemu Guest Agent is enabled but no IP config is found"
	errorGuestAgentNoIPv4Summary string = "Qemu Guest Agent is enabled but no IPv4 address is found"
	errorGuestAgentNoIPv6Summary string = "Qemu Guest Agent is enabled but no IPv6 address is found"
)

func parseCloudInitInterface(ipConfig pxapi.CloudInitNetworkConfig, ciCustom, skipIPv4, skipIPv6 bool) (conn connectionInfo) {
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

func (conn connectionInfo) agentDiagnostics() diag.Diagnostics {
	if conn.IPs.IPv4 == "" {
		if conn.IPs.IPv6 == "" {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  errorGuestAgentNoIPSummary,
				Detail:   "Qemu Guest Agent is enabled in your configuration but no IP address was found before the time ran out, increasing 'agent_timeout' could resolve this issue."}}
		}
		if !conn.SkipIPv4 {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  errorGuestAgentNoIPv4Summary,
				Detail:   "Qemu Guest Agent is enabled in your configuration but no IPv4 address was found before the time ran out, increasing 'agent_timeout' could resolve this issue. To suppress this warning set 'skip_ipv4' to true."}}
		}
		return diag.Diagnostics{}
	}
	if conn.IPs.IPv6 == "" && !conn.SkipIPv6 {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errorGuestAgentNoIPv6Summary,
			Detail:   "Qemu Guest Agent is enabled in your configuration but no IPv6 address was found before the time ran out, increasing 'agent_timeout' could resolve this issue. To suppress this warning set 'skip_ipv6' to true."}}
	}
	return diag.Diagnostics{}
}

func (conn connectionInfo) hasRequiredIP() bool {
	if conn.IPs.IPv4 == "" && !conn.SkipIPv4 || conn.IPs.IPv6 == "" && !conn.SkipIPv6 {
		return false
	}
	return true
}

func (conn connectionInfo) parsePrimaryIPs(interfaces []pxapi.AgentNetworkInterface, mac string) connectionInfo {
	lowerCaseMac := strings.ToLower(mac)
	for _, iFace := range interfaces {
		if iFace.MacAddress.String() == lowerCaseMac {
			for _, addr := range iFace.IpAddresses {
				if addr.IsGlobalUnicast() {
					if addr.To4() != nil {
						if conn.IPs.IPv4 == "" {
							conn.IPs.IPv4 = addr.String()
						}
					} else {
						if conn.IPs.IPv6 == "" {
							conn.IPs.IPv6 = addr.String()
						}
					}
				}
			}
		}
	}
	return conn
}
