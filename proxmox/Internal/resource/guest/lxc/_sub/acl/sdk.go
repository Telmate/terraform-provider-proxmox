package acl

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
)

func SDK(acl string) *pveSDK.TriBool {
	switch acl {
	case flagTrue:
		return util.Pointer(pveSDK.TriBoolTrue)
	case flagFalse:
		return util.Pointer(pveSDK.TriBoolFalse)
	default:
		return util.Pointer(pveSDK.TriBoolNone)
	}
}
