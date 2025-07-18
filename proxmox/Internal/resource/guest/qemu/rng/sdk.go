package rng

import (
	"time"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.VirtIoRNG {
	if v := d.Get(Root).([]any); len(v) > 0 && v[0] != nil {
		settings := v[0].(map[string]any)
		var source pveSDK.EntropySource
		_ = source.Parse(settings[schemaSource].(string))
		return &pveSDK.VirtIoRNG{
			Limit:  util.Pointer(uint(settings[schemaLimit].(int))),
			Period: util.Pointer(time.Duration(settings[schemaPeriod].(int)) * time.Millisecond),
			Source: &source}
	}
	return &pveSDK.VirtIoRNG{Delete: true}
}
