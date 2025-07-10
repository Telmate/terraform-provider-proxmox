package memory

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(memory *pveSDK.LxcMemory, d *schema.ResourceData) {
	d.Set(Root, int(*memory))
}
