package operatingsystem

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(os pveSDK.OperatingSystem, d *schema.ResourceData) {
	d.Set(Root, string(os))
}
