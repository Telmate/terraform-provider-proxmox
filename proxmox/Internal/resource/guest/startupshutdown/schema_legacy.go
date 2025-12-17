package startupshutdown

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const LegacyRoot = "startup"

func LegacySchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Startup order of the VM",
		Deprecated:    "Use " + Root + " instead",
		ConflictsWith: []string{Root},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return printLegacyStartupShutdown_Unsafe(parseLegacyStartupShutdown(old)) == printLegacyStartupShutdown_Unsafe(parseLegacyStartupShutdown(new))
		}}
}
