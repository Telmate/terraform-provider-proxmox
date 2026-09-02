package powerstate

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveSDK.PowerState, legacy bool, d *schema.ResourceData) {
	if legacy {
		if _, ok := d.GetOk(LegacyRoot); ok {
			terraformLegacy(config,d)
			return
		}
		terraformLegacyClear(d)
	}
	d.Set(Root, terraform(config))
}

func terraform(config pveSDK.PowerState) string {
	switch config {
	case pveSDK.PowerStateRunning:
		return enumRunning
	case pveSDK.PowerStateStopped:
		return enumStopped
	default:
		return ""
	}
}
