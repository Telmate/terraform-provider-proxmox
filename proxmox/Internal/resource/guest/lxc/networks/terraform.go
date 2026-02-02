package networks

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(networks pveSDK.LxcNetworks, d *schema.ResourceData) error {
	if tfConfig, ok := d.GetOk(RootNetwork); ok {
		devices, err := terraformNetwork(networks, tfConfig.([]any))
		if err != nil {
			return err
		}
		d.Set(RootNetwork, devices)
	} else {
		d.Set(RootNetworks, terraformNetworks(networks))
	}
	return nil
}
