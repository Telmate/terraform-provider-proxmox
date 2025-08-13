package mounts

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
)

func sdkMounts(privileged bool, schema map[string]any) pveSDK.LxcMounts {
	config := make(pveSDK.LxcMounts, len(schema))
	for k, v := range schema {
		tmpID, _ := strconv.ParseUint(k[len(prefixSchemaID):], 10, 64)
		schemaArray := v.([]any)
		if len(schemaArray) == 0 {
			config[pveSDK.LxcMountID(tmpID)] = pveSDK.LxcMount{Detach: true}
			continue
		}
		schemaMap := schemaArray[0].(map[string]any)
		if v, ok := schemaMap[schemaDataMount].([]any); ok && len(v) == 1 && v[0] != nil {
			settings := v[0].(map[string]any)
			var quota *bool
			if privileged {
				quota = util.Pointer(settings[schemaQuota].(bool))
			}
			config[pveSDK.LxcMountID(tmpID)] = pveSDK.LxcMount{DataMount: &pveSDK.LxcDataMount{
				ACL:    acl.SDK(settings[schemaACL].(string)),
				Backup: util.Pointer(settings[schemaBackup].(bool)),
				Options: &pveSDK.LxcMountOptions{
					Discard:  util.Pointer(settings[schemaOptionsDiscard].(bool)),
					LazyTime: util.Pointer(settings[schemaOptionsLazyTime].(bool)),
					NoATime:  util.Pointer(settings[schemaOptionsNoATime].(bool)),
					NoDevice: util.Pointer(settings[schemaOptionsNoDevice].(bool)),
					NoExec:   util.Pointer(settings[schemaOptionsNoExec].(bool)),
					NoSuid:   util.Pointer(settings[schemaOptionsNoSuid].(bool))},
				Path:            util.Pointer(pveSDK.LxcMountPath(settings[schemaGuestPath].(string))),
				Quota:           quota,
				ReadOnly:        util.Pointer(settings[schemaReadOnly].(bool)),
				Replicate:       util.Pointer(settings[schemaReplicate].(bool)),
				SizeInKibibytes: util.Pointer(pveSDK.LxcMountSize(size.Parse_Unsafe(settings[schemaSize].(string)))),
				Storage:         util.Pointer(settings[schemaStorage].(string))}}
			continue
		}
		if v, ok := schemaMap[schemaBindMount].([]any); ok && len(v) == 1 && v[0] != nil {
			settings := v[0].(map[string]any)
			config[pveSDK.LxcMountID(tmpID)] = pveSDK.LxcMount{BindMount: &pveSDK.LxcBindMount{
				GuestPath: util.Pointer(pveSDK.LxcMountPath(settings[schemaGuestPath].(string))),
				HostPath:  util.Pointer(pveSDK.LxcHostPath(settings[schemaHostPath].(string))),
				Options: &pveSDK.LxcMountOptions{
					Discard:  util.Pointer(settings[schemaOptionsDiscard].(bool)),
					LazyTime: util.Pointer(settings[schemaOptionsLazyTime].(bool)),
					NoATime:  util.Pointer(settings[schemaOptionsNoATime].(bool)),
					NoDevice: util.Pointer(settings[schemaOptionsNoDevice].(bool)),
					NoExec:   util.Pointer(settings[schemaOptionsNoExec].(bool)),
					NoSuid:   util.Pointer(settings[schemaOptionsNoSuid].(bool))},
				ReadOnly:  util.Pointer(settings[schemaReadOnly].(bool)),
				Replicate: util.Pointer(settings[schemaReplicate].(bool))}}
		}
	}
	return config
}
