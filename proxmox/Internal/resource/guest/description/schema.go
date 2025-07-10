package description

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "description"

	LegacyQemu = "desc"

	defaultRoot = "Managed by Terraform."
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  defaultRoot,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return strings.TrimSpace(old) == strings.TrimSpace(new)
		}}
}

func LegacySchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Deprecated:    "Use '" + Root + "' instead.",
		ConflictsWith: []string{Root},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return strings.TrimSpace(old) == strings.TrimSpace(new)
		}}
}
