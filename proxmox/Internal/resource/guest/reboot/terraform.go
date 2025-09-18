package reboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func SetRequired(v bool, d *schema.ResourceData) { d.Set(RootRequired, v) }
