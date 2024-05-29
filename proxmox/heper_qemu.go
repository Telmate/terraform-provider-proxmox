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

func parseCloudInitInterface(ipConfig string, skipIPv4, skipIPv6 bool) (conn connectionInfo) {
	conn.SkipIPv4 = skipIPv4
	conn.SkipIPv6 = skipIPv6
	var IPv4Set, IPv6Set bool
	for _, e := range strings.Split(ipConfig, ",") {
		if len(e) < 4 {
			continue
		}
		if e[:3] == "ip=" {
			IPv4Set = true
			splitCIDR := strings.Split(e[3:], "/")
			if len(splitCIDR) == 2 {
				conn.IPs.IPv4 = splitCIDR[0]
			}
		}
		if e[:4] == "ip6=" {
			IPv6Set = true
			splitCIDR := strings.Split(e[4:], "/")
			if len(splitCIDR) == 2 {
				conn.IPs.IPv6 = splitCIDR[0]
			}
		}
	}
	if !IPv4Set && conn.IPs.IPv4 == "" {
		conn.SkipIPv4 = true
	}
	if !IPv6Set && conn.IPs.IPv6 == "" {
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
