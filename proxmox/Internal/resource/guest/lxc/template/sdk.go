package template

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.LxcTemplate {
	v, ok := d.GetOk(Root)
	if !ok {
		return nil
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return nil
	}
	settings, ok := vv[0].(map[string]any)
	if !ok {
		return nil
	}
	return &pveSDK.LxcTemplate{
		File:    settings[schemaFile].(string),
		Storage: settings[schemaStorage].(string)}
}
