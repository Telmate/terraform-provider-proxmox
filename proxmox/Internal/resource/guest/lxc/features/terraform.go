package features

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(features *pveSDK.LxcFeatures, d *schema.ResourceData) {
	if features == nil {
		return
	}
	var settings map[string]any
	if features.Privileged != nil {
		settings = map[string]any{
			schemaPrivileged: []map[string]any{{
				schemaCreateDeviceNodes: features.Privileged.CreateDeviceNodes,
				schemaFUSE:              features.Privileged.FUSE,
				schemaNFS:               features.Privileged.NFS,
				schemaNesting:           features.Privileged.Nesting,
				schemaSMB:               features.Privileged.SMB}}}
	} else if features.Unprivileged != nil {
		settings = map[string]any{
			schemaUnprivileged: []map[string]any{{
				schemaCreateDeviceNodes: features.Unprivileged.CreateDeviceNodes,
				schemaFUSE:              features.Unprivileged.FUSE,
				schemaNesting:           features.Unprivileged.Nesting,
				schemeKeyCtl:            features.Unprivileged.KeyCtl}}}
	}
	d.Set(Root, []map[string]any{settings})
}
