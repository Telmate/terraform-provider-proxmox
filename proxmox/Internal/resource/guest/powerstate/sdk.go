package powerstate

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.PowerState {
	switch d.Get(Root).(string) {
	case enumRunning:
		state := pveSDK.PowerStateRunning
		return &state
	case enumStopped:
		state := pveSDK.PowerStateStopped
		return &state
	default:
		return nil
	}
}
