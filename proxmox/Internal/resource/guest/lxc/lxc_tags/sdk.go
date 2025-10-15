package lxc_tags

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.Tags {
	if v, ok := d.GetOk(Root); ok {
		rawTags := v.(*schema.Set).List()
		tags := make(pveSDK.Tags, len(rawTags))
		for i := range rawTags {
			tags[i] = pveSDK.Tag(rawTags[i].(string))
		}
		return &tags
	}
	return util.Pointer(pveSDK.Tags{})
}
