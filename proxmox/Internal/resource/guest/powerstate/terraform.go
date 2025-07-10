package powerstate

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(state pveSDK.PowerState, d *schema.ResourceData) {
	switch state {
	case pveSDK.PowerStateRunning:
		d.Set(Root, enumRunning)
	case pveSDK.PowerStateStopped:
		d.Set(Root, enumStopped)
	default:
		d.Set(Root, "")
	}
}
