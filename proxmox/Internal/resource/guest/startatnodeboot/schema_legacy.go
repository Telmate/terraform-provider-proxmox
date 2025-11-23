package startatnodeboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const LegacyRoot = "onboot"

func LegacySchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		Description:   "VM autostart on boot",
		Deprecated:    "Use " + Root + " instead",
		ConflictsWith: []string{Root},
		Optional:      true}
}
