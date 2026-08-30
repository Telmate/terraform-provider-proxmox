package wait

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const Root = "additional_wait"

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     5,
		Description: "Value in second to wait after some operations, useful if system is not fast or during I/O intensive parallel terraform tasks",
	}
}

func GetDuration(d *schema.ResourceData) time.Duration {
	return time.Duration(d.Get(Root).(int)) * time.Second
}
