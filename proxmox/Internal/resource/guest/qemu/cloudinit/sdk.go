package cloudinit

import (
	"strconv"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/pve/dns/nameservers"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/sshkeys"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.CloudInit {
	ci := pveSDK.CloudInit{
		Custom: sdkCloudInitCustom(d.Get(RootCustom).(string)),
		DNS: &pveSDK.GuestDNS{
			SearchDomain: util.Pointer(d.Get(RootSearchDomain).(string)),
			NameServers:  nameservers.Split(d.Get(RootNameServers).(string))},
		NetworkInterfaces: pveSDK.CloudInitNetworkInterfaces{},
		PublicSSHkeys:     sshkeys.SDK(d),
		UpgradePackages:   util.Pointer(d.Get(RootUpgrade).(bool)),
		UserPassword:      util.Pointer(d.Get(RootPassword).(string)),
		Username:          util.Pointer(d.Get(RootUser).(string))}
	for i := 0; i < 16; i++ {
		ci.NetworkInterfaces[pveSDK.QemuNetworkInterfaceID(i)] = sdkCloudInitNetworkConfig(d.Get(prefixNetworkConfig + strconv.Itoa(i)).(string))
	}
	return &ci
}

func sdkCloudInitCustom(settings string) *pveSDK.CloudInitCustom {
	var meta, network, user, vendor pveSDK.CloudInitSnippet
	params := splitStringOfSettings(settings)
	if v, ok := params["meta"]; ok {
		meta = sdkCloudInitSnippet(v)
	}
	if v, ok := params["network"]; ok {
		network = sdkCloudInitSnippet(v)
	}
	if v, ok := params["user"]; ok {
		user = sdkCloudInitSnippet(v)
	}
	if v, ok := params["vendor"]; ok {
		vendor = sdkCloudInitSnippet(v)
	}
	return &pveSDK.CloudInitCustom{
		Meta:    &meta,
		Network: &network,
		User:    &user,
		Vendor:  &vendor}
}

func sdkCloudInitSnippet(param string) pveSDK.CloudInitSnippet {
	file := strings.SplitN(param, ":", 2)
	if len(file) == 2 {
		return pveSDK.CloudInitSnippet{
			Storage:  file[0],
			FilePath: pveSDK.CloudInitSnippetPath(file[1])}
	}
	return pveSDK.CloudInitSnippet{}
}

func sdkCloudInitNetworkConfig(param string) pveSDK.CloudInitNetworkConfig {
	config := pveSDK.CloudInitNetworkConfig{
		IPv4: &pveSDK.CloudInitIPv4Config{
			Address: util.Pointer(pveSDK.IPv4CIDR("")),
			DHCP:    false,
			Gateway: util.Pointer(pveSDK.IPv4Address(""))},
		IPv6: &pveSDK.CloudInitIPv6Config{
			Address: util.Pointer(pveSDK.IPv6CIDR("")),
			DHCP:    false,
			Gateway: util.Pointer(pveSDK.IPv6Address("")),
			SLAAC:   false}}
	params := splitStringOfSettings(param)
	if v, isSet := params["ip"]; isSet {
		if v == "dhcp" {
			config.IPv4.DHCP = true
		} else {
			*config.IPv4.Address = pveSDK.IPv4CIDR(v)
		}
	}
	if v, isSet := params["gw"]; isSet {
		*config.IPv4.Gateway = pveSDK.IPv4Address(v)
	}
	if v, isSet := params["ip6"]; isSet {
		switch v {
		case "dhcp":
			config.IPv6.DHCP = true
		case "auto":
			config.IPv6.SLAAC = true
		default:
			*config.IPv6.Address = pveSDK.IPv6CIDR(v)
		}
	}
	if v, isSet := params["gw6"]; isSet {
		*config.IPv6.Gateway = pveSDK.IPv6Address(v)
	}
	return config
}
