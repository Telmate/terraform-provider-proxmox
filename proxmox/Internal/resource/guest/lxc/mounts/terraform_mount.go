package mounts

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
)

func terraformMount(config pveSDK.LxcMounts, tfConfig []any) []map[string]any {
	mounts := make([]map[string]any, len(config))
	var index int
	for i := range pveSDK.LxcMountID(maximumID) {
		v, ok := config[i]
		if !ok {
			continue
		}
		var params map[string]any
		if v.DataMount != nil {
			params = map[string]any{
				schemaACL:       acl.Terraform(v.DataMount.ACL),
				schemaBackup:    *v.DataMount.Backup,
				schemaGuestPath: v.DataMount.Path.String(),
				schemaQuota:     *v.DataMount.Quota,
				schemaReadOnly:  *v.DataMount.ReadOnly,
				schemaReplicate: *v.DataMount.Replicate,
				schemaSize:      size.String(int64(*v.DataMount.SizeInKibibytes)),
				schemaStorage:   *v.DataMount.Storage}
			terraformSetOptions(params, v.DataMount.Options)
		} else if v.BindMount != nil {
			params = map[string]any{
				schemaGuestPath: v.BindMount.GuestPath.String(),
				schemaHostPath:  v.BindMount.HostPath.String(),
				schemaReadOnly:  *v.BindMount.ReadOnly,
				schemaReplicate: *v.BindMount.Replicate}
			terraformSetOptions(params, v.BindMount.Options)
		}
		params[schemaID] = prefixSchemaID + strconv.Itoa(int(i))
		mounts[index] = params
		index++
	}
	return mounts
}
