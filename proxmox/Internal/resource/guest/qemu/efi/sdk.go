package efi

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.EfiDisk {
	v, ok := d.GetOk(Root)
	if !ok {
		return nil
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return nil
	}
	if settings, ok := vv[0].(map[string]any); ok {
		if storage := settings[schemaStorage].(string); storage != "" {
			disk := pveSDK.EfiDisk{
				Format:          new(defaultFormat),
				PreEnrolledKeys: new(defaultPreEnrolledKeys),
				Storage:         new(pveSDK.StorageName(storage)),
				Type:            new(defaultEfiType),
			}
			if v, ok := settings[schemaEfiType].(string); ok && v != "" {
				*disk.Type = pveSDK.EfiDiskType(v)
			}
			if v, ok := settings[schemaFormat].(string); ok && v != "" {
				*disk.Format = pveSDK.QemuDiskFormat(v)
			}
			if v, ok := settings[schemaPreEnrolledKeys].(bool); ok {
				*disk.PreEnrolledKeys = v
			}
			return &disk
		}
		return &pveSDK.EfiDisk{Delete: true}
	}
	return nil
}
