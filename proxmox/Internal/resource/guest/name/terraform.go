package name

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform_Unsafe(name *pveSDK.GuestName, d *schema.ResourceData) {
	if name == nil {
		return
	}
	d.Set(Root, string(*name))
}
