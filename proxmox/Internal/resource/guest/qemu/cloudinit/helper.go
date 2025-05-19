package cloudinit

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NeedsReboot(ci *pveSDK.CloudInit, d *schema.ResourceData) bool {
	if ci != nil && ci.UpgradePackages != nil && *ci.UpgradePackages && d.HasChange(RootUpgrade) {
		return true
	}
	return d.HasChanges(
		RootUser,
		RootSearchDomain,
		RootPassword,
		RootCustom,
		RootNameServers,
		RootNetworkConfig0,
		RootNetworkConfig1,
		RootNetworkConfig2,
		RootNetworkConfig3,
		RootNetworkConfig4,
		RootNetworkConfig5,
		RootNetworkConfig6,
		RootNetworkConfig7,
		RootNetworkConfig8,
		RootNetworkConfig9,
		RootNetworkConfig10,
		RootNetworkConfig11,
		RootNetworkConfig12,
		RootNetworkConfig13,
		RootNetworkConfig14,
		RootNetworkConfig15)
}

func splitStringOfSettings(settings string) map[string]string {
	settingValuePairs := strings.Split(settings, ",")
	settingMap := map[string]string{}
	for _, e := range settingValuePairs {
		keyValuePair := strings.SplitN(e, "=", 2)
		var value string
		if len(keyValuePair) == 2 {
			value = keyValuePair[1]
		}
		settingMap[keyValuePair[0]] = value
	}
	return settingMap
}

func trimNameServers(nameServers string) string {
	return strings.ReplaceAll(strings.ReplaceAll(nameServers, " ", ""), ",", "")
}
