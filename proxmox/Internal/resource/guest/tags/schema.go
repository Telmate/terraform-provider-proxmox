package tags

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "tags"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf("expected a string, got: %s", i)
			}
			for _, e := range *split(v) {
				if err := e.Validate(); err != nil {
					return diag.Errorf("tag validation failed: %s", err)
				}
			}
			return nil
		},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return toString(sortArray(removeDuplicates(split(old)))) == toString(sortArray(removeDuplicates(split(new))))
		},
	}
}
