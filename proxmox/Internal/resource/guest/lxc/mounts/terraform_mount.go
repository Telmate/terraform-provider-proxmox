package mounts

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
)

func terraformMount(config pveSDK.LxcMounts, tfConfig []any) []map[string]any {
	mounts := make([]map[string]any, len(config))
	configMap := make(map[string]map[string]any, len(tfConfig))
	for i := range tfConfig {
		subMap := tfConfig[i].(map[string]any)
		configMap[subMap[schemaID].(string)] = subMap
	}
	var index int
	for i := range pveSDK.LxcMountID(maximumID) {
		v, ok := config[i]
		if !ok {
			continue
		}
		id := prefixSchemaID + strconv.Itoa(int(i))
		var params map[string]any
		if v.DataMount != nil {
			var quota bool
			if v.DataMount.Quota != nil {
				quota = *v.DataMount.Quota
			}
			params = map[string]any{
				schemaACL:       acl.Terraform(v.DataMount.ACL),
				schemaBackup:    *v.DataMount.Backup,
				schemaGuestPath: v.DataMount.Path.String(),
				schemaQuota:     quota,
				schemaReadOnly:  *v.DataMount.ReadOnly,
				schemaReplicate: *v.DataMount.Replicate,
				schemaSize:      size.String(int64(*v.DataMount.SizeInKibibytes)),
				schemaStorage:   *v.DataMount.Storage,
				schemaType:      typeDataMount}
			terraformSetOptions(params, v.DataMount.Options)
		} else if v.BindMount != nil {
			localMap := configMap[id]
			params = map[string]any{
				schemaBackup:    localMap[schemaBackup].(bool),
				schemaGuestPath: v.BindMount.GuestPath.String(),
				schemaHostPath:  v.BindMount.HostPath.String(),
				schemaReadOnly:  *v.BindMount.ReadOnly,
				schemaReplicate: *v.BindMount.Replicate,
				schemaType:      typeBindMount}
			terraformSetOptions(params, v.BindMount.Options)
		}
		params[schemaID] = id
		mounts[index] = params
		index++
	}
	return mounts
}
