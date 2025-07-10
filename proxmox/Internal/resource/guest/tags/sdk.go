package tags

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.Tags {
	if v, ok := d.GetOk(Root); ok {
		return removeDuplicates(split(v.(string)))
	}
	return util.Pointer(pveSDK.Tags{})
}
