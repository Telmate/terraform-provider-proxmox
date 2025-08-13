package mounts

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
)

func terraformMounts(config pveSDK.LxcMounts) []any {
	mapParams := make(map[string]any, mountsAmount)
	for k, v := range config {
		index := prefixSchemaID + strconv.Itoa(int(k))
		settings := make(map[string]any, 1)
		if v.DataMount != nil {
			dataMount := map[string]any{
				schemaACL:       acl.Terraform(v.DataMount.ACL),
				schemaBackup:    *v.DataMount.Backup,
				schemaGuestPath: v.DataMount.Path.String(),
				schemaReadOnly:  *v.DataMount.ReadOnly,
				schemaReplicate: *v.DataMount.Replicate,
				schemaSize:      size.String(int64(*v.DataMount.SizeInKibibytes)),
				schemaStorage:   *v.DataMount.Storage}
			if v.DataMount.Quota != nil {
				dataMount[schemaQuota] = *v.DataMount.Quota
			}
			terraformSetOptions(dataMount, v.DataMount.Options)
			settings[schemaDataMount] = []map[string]any{dataMount}
		} else if v.BindMount != nil {
			bindMount := map[string]any{
				schemaHostPath:  v.BindMount.HostPath.String(),
				schemaGuestPath: v.BindMount.GuestPath.String(),
				schemaReadOnly:  *v.BindMount.ReadOnly,
				schemaReplicate: *v.BindMount.Replicate}
			terraformSetOptions(bindMount, v.BindMount.Options)
			settings[schemaBindMount] = []any{bindMount}
		}
		mapParams[index] = []any{settings}
	}
	return []any{mapParams}
}

func terraformSetOptions(params map[string]any, options *pveSDK.LxcMountOptions) {
	if options != nil {
		params[schemaOptionsDiscard] = *options.Discard
		params[schemaOptionsLazyTime] = *options.LazyTime
		params[schemaOptionsNoATime] = *options.NoATime
		params[schemaOptionsNoDevice] = *options.NoDevice
		params[schemaOptionsNoExec] = *options.NoExec
		params[schemaOptionsNoSuid] = *options.NoSuid
	} else {
		params[schemaOptionsDiscard] = false
		params[schemaOptionsLazyTime] = false
		params[schemaOptionsNoATime] = false
		params[schemaOptionsNoDevice] = false
		params[schemaOptionsNoExec] = false
		params[schemaOptionsNoSuid] = false
	}
}
