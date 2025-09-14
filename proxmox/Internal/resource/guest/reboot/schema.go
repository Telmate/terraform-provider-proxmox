package reboot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "automatic_reboot"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Automatically reboot the guest system if any of the modified parameters requires a reboot to take effect.",
		Optional:    true,
		Default:     true}
}
