package cloudinit

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/pve/dns/nameservers"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/sshkeys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config *pveSDK.CloudInit, d *schema.ResourceData) {
	// we purposely use the password from the terraform config here
	// because the proxmox api will always return "**********" leading to diff issues
	d.Set(RootPassword, d.Get(RootPassword).(string))

	d.Set(RootUser, config.Username)
	if config.Custom != nil {
		d.Set(RootCustom, config.Custom.String())
	}
	if config.DNS != nil {
		d.Set(RootSearchDomain, config.DNS.SearchDomain)
		d.Set(RootNameServers, nameservers.String(config.DNS.NameServers))
	}
	for i := pveSDK.QemuNetworkInterfaceID(0); i < 16; i++ {
		if v, isSet := config.NetworkInterfaces[i]; isSet {
			d.Set(prefixNetworkConfig+strconv.Itoa(int(i)), terraformCloudInitNetworkConfig(v))
		}
	}
	if config.PublicSSHkeys != nil {
		sshkeys.Terraform(*config.PublicSSHkeys, d)
	}
	if config.UpgradePackages != nil {
		d.Set(RootUpgrade, *config.UpgradePackages)
	}
}

func terraformCloudInitNetworkConfig(config pveSDK.CloudInitNetworkConfig) string {
	if config.IPv4 != nil {
		if config.IPv6 != nil {
			return config.IPv4.String() + "," + config.IPv6.String()
		} else {
			return config.IPv4.String()
		}
	} else {
		if config.IPv6 != nil {
			return config.IPv6.String()
		}
	}
	return ""
}
