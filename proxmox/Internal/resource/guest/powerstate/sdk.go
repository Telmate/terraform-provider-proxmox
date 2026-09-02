package powerstate

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	LegacyFalse  LegacyEnum = 0
	LegacyCreate LegacyEnum = 1
	LegacyUpdate LegacyEnum = 2
)

type LegacyEnum uint8

func SDK(legacy LegacyEnum, d *schema.ResourceData) *pveSDK.PowerState {
	v, ok := d.GetOk(Root)
	if legacy > LegacyFalse {
		if _, okLegacy := d.GetOk(LegacyRoot); okLegacy || !ok {
			return sdkLegacy(d)
		}
	}
	switch v.(string) {
	case enumRunning:
		return util.Pointer(pveSDK.PowerStateRunning)
	case enumStopped:
		return util.Pointer(pveSDK.PowerStateStopped)
	default:
		return nil
	}
}
