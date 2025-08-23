package mounts

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func sdkMount(privileged bool, schema []any) (pveSDK.LxcMounts, diag.Diagnostics) {
	config := sdkDefaults()
	diags := diag.Diagnostics{}
	for _, e := range schema {
		schemaMap := e.(map[string]any)
		rawID, _ := strconv.ParseUint(schemaMap[schemaID].(string)[2:], 10, 64)
		id := pveSDK.LxcMountID(rawID)
		if v := config[id]; !v.Detach {
			return nil, append(diags, diag.Diagnostic{
				Summary:  "duplicate mount " + prefixSchemaID + id.String(),
				Severity: diag.Error})
		}
		options := pveSDK.LxcMountOptions{
			Discard:  util.Pointer(schemaMap[schemaOptionsDiscard].(bool)),
			LazyTime: util.Pointer(schemaMap[schemaOptionsLazyTime].(bool)),
			NoATime:  util.Pointer(schemaMap[schemaOptionsNoATime].(bool)),
			NoDevice: util.Pointer(schemaMap[schemaOptionsNoDevice].(bool)),
			NoExec:   util.Pointer(schemaMap[schemaOptionsNoExec].(bool)),
			NoSuid:   util.Pointer(schemaMap[schemaOptionsNoSuid].(bool))}
		switch schemaMap[schemaType].(string) {
		case typeBindMount:
			var hostPath string
			if v := schemaMap[schemaHostPath].(string); v != "" {
				hostPath = v
			} else {
				return nil, append(diags, diag.Diagnostic{
					Summary:  schemaHostPath + " is required for " + typeBindMount + " mount",
					Severity: diag.Error})
			}
			// warnings for unused settings
			if schemaMap[schemaACL].(string) != acl.Default {
				diags = append(diags, warning(schemaACL, typeBindMount))
			}
			if !schemaMap[schemaBackup].(bool) {
				diags = append(diags, warning(schemaBackup, typeBindMount))
			}
			if schemaMap[schemaQuota].(bool) {
				diags = append(diags, warning(schemaQuota, typeBindMount))
			}
			if schemaMap[schemaSize].(string) != "" {
				diags = append(diags, warning(schemaSize, typeBindMount))
			}
			if schemaMap[schemaStorage].(string) != "" {
				diags = append(diags, warning(schemaStorage, typeBindMount))
			}
			// configure mount
			config[id] = pveSDK.LxcMount{BindMount: &pveSDK.LxcBindMount{
				GuestPath: util.Pointer(pveSDK.LxcMountPath(schemaMap[schemaGuestPath].(string))),
				HostPath:  util.Pointer(pveSDK.LxcHostPath(hostPath)),
				Options:   &options,
				ReadOnly:  util.Pointer(schemaMap[schemaReadOnly].(bool)),
				Replicate: util.Pointer(schemaMap[schemaReplicate].(bool))}}
		case typeDataMount:
			rawSize := schemaMap[schemaSize].(string)
			if rawSize == "" {
				return nil, append(diags, diag.Diagnostic{
					Summary:  schemaSize + " is required for " + typeDataMount + " mount",
					Severity: diag.Error})
			}
			storage := schemaMap[schemaStorage].(string)
			if storage == "" {
				return nil, append(diags, diag.Diagnostic{
					Summary:  schemaStorage + " is required for " + typeDataMount + " mount",
					Severity: diag.Error})
			}
			// warnings for unused settings
			if schemaMap[schemaHostPath].(string) != "" {
				diags = append(diags, warning(schemaHostPath, typeDataMount))
			}

			// configure mount
			var quota *bool
			if privileged {
				quota = util.Pointer(schemaMap[schemaQuota].(bool))
			}
			config[id] = pveSDK.LxcMount{DataMount: &pveSDK.LxcDataMount{
				ACL:             acl.SDK(schemaMap[schemaACL].(string)),
				Backup:          util.Pointer(schemaMap[schemaBackup].(bool)),
				Options:         &options,
				Path:            util.Pointer(pveSDK.LxcMountPath(schemaMap[schemaGuestPath].(string))),
				Quota:           quota,
				ReadOnly:        util.Pointer(schemaMap[schemaReadOnly].(bool)),
				Replicate:       util.Pointer(schemaMap[schemaReplicate].(bool)),
				SizeInKibibytes: util.Pointer(pveSDK.LxcMountSize(size.Parse_Unsafe(rawSize))),
				Storage:         &storage}}
		}
	}
	return config, diags
}

func warning(key, kind string) diag.Diagnostic {
	return diag.Diagnostic{
		Summary:  key + " is not used when " + schemaType + " is " + kind,
		Severity: diag.Warning}
}
