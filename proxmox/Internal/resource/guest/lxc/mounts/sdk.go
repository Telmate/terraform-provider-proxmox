package mounts

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(privileged bool, d *schema.ResourceData) (pveSDK.LxcMounts, diag.Diagnostics) {

	if v, ok := d.GetOk(RootMount); ok {
		return sdkMount(privileged, v.([]any))
	} else {
		if v := d.Get(RootMounts).([]any); len(v) == 1 {
			if subSchema, ok := v[0].(map[string]any); ok {
				return sdkMounts(privileged, subSchema), nil
			}
		}
	}
	return sdkDefaults(), nil
}

func sdkDefaults() pveSDK.LxcMounts {
	config := make(pveSDK.LxcMounts, mountsAmount)
	for i := range pveSDK.LxcMountID(maximumID) {
		config[i] = pveSDK.LxcMount{Detach: true}
	}
	return config
}
