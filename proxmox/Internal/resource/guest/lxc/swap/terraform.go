package swap

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(swap *pveSDK.LxcSwap, d *schema.ResourceData) {
	d.Set(Root, int(*swap))
}
