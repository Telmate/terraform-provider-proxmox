package tags

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(tags *pveSDK.Tags, d *schema.ResourceData) {
	d.Set(Root, toString(tags))
}
