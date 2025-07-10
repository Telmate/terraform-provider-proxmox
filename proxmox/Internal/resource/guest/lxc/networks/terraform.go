package networks

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(networks pveSDK.LxcNetworks, d *schema.ResourceData) {
	if tfConfig, ok := d.GetOk(RootNetwork); ok {
		d.Set(RootNetwork, terraformNetwork(networks, tfConfig.([]any)))
	} else {
		d.Set(RootNetworks, terraformNetworks(networks))
	}
}
