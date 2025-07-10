package architecture

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(architecture pveSDK.CpuArchitecture, d *schema.ResourceData) {
	d.Set(Root, architecture.String())
}
