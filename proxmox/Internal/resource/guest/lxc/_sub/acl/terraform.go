package acl

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
)

func Terraform(acl *pveSDK.TriBool) string {
	switch *acl {
	case pveSDK.TriBoolTrue:
		return flagTrue
	case pveSDK.TriBoolFalse:
		return flagFalse
	default:
		return flagDefault
	}
}
