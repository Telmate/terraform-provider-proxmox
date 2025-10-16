package lxc_tags

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(tags *pveSDK.Tags, d *schema.ResourceData) {
	if tags == nil {
		d.Set(Root, nil)
		return
	}
	tagSet := make([]any, len(*tags))
	for i := range *tags {
		tagSet[i] = string((*tags)[i])
	}
	d.Set(Root, tagSet)
}
