package features

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(Privileged bool, d *schema.ResourceData) *pveSDK.LxcFeatures {
	schemaItem, ok := d.GetOk(Root)
	if !ok {
		return defaults(Privileged)
	}
	schemaFeatures, ok := schemaItem.([]any)
	if ok && len(schemaFeatures) != 1 {
		return defaults(Privileged)
	}
	settings, ok := schemaFeatures[0].(map[string]any)
	if !ok {
		return defaults(Privileged)
	}
	if v, ok := settings[schemaPrivileged].([]any); ok && len(v) == 1 && v[0] != nil {
		features := v[0].(map[string]any)
		return &pveSDK.LxcFeatures{
			Privileged: &pveSDK.PrivilegedFeatures{
				CreateDeviceNodes: util.Pointer(features[schemaCreateDeviceNodes].(bool)),
				FUSE:              util.Pointer(features[schemaFUSE].(bool)),
				NFS:               util.Pointer(features[schemaNFS].(bool)),
				Nesting:           util.Pointer(features[schemaNesting].(bool)),
				SMB:               util.Pointer(features[schemaSMB].(bool))}}
	}
	if v, ok := settings[schemaUnprivileged].([]any); ok && len(v) == 1 && v[0] != nil {
		features := v[0].(map[string]any)
		return &pveSDK.LxcFeatures{
			Unprivileged: &pveSDK.UnprivilegedFeatures{
				CreateDeviceNodes: util.Pointer(features[schemaCreateDeviceNodes].(bool)),
				FUSE:              util.Pointer(features[schemaFUSE].(bool)),
				KeyCtl:            util.Pointer(features[schemeKeyCtl].(bool)),
				Nesting:           util.Pointer(features[schemaNesting].(bool))}}
	}
	return defaults(Privileged)
}

func defaults(Privileged bool) *pveSDK.LxcFeatures {
	if Privileged {
		return &pveSDK.LxcFeatures{
			Privileged: &pveSDK.PrivilegedFeatures{
				CreateDeviceNodes: util.Pointer(defaultCreateDeviceNodes),
				FUSE:              util.Pointer(defaultFUSE),
				NFS:               util.Pointer(defaultNFS),
				Nesting:           util.Pointer(defaultNesting),
				SMB:               util.Pointer(defaultSMB)}}
	}
	return &pveSDK.LxcFeatures{
		Unprivileged: &pveSDK.UnprivilegedFeatures{
			CreateDeviceNodes: util.Pointer(defaultCreateDeviceNodes),
			FUSE:              util.Pointer(defaultFUSE),
			KeyCtl:            util.Pointer(defaultKeyCtl),
			Nesting:           util.Pointer(defaultNesting)}}
}
