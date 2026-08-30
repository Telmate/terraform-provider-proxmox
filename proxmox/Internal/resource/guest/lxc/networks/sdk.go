package networks

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(version pveSDK.EncodedVersion, d *schema.ResourceData) (pveSDK.LxcNetworks, diag.Diagnostics) {
	if v, ok := d.GetOk(RootNetwork); ok { // network
		return sdkNetwork(version, v.([]any))
	} else if v := d.Get(RootNetworks).([]any); len(v) == 1 { // networks
		if subSchema, ok := v[0].(map[string]any); ok {
			return sdkNetworks(version, subSchema), nil
		}
	}
	// Defaults
	config := make(pveSDK.LxcNetworks, NetworksAmount)
	for i := range pveSDK.LxcNetworkID(maximumID) {
		config[i] = pveSDK.LxcNetwork{Delete: true}
	}
	return config, nil
}
