package proxmox

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTimeouts() *schema.ResourceTimeout {
	// resourceCreateTimeout := defaultTimeout
	// resourceReadTimeout := 600
	// resourceUpdateTimeout := 600
	// resourceDeleteTimeout := 1200

	// if v, ok := os.LookupEnv("PM_TIMEOUT"); ok {
	// 	resourceCreateTimeout, _ = strconv.Atoi(v)
	// }

	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(20 * time.Minute),
		Read:    schema.DefaultTimeout(20 * time.Minute),
		Update:  schema.DefaultTimeout(20 * time.Minute),
		Delete:  schema.DefaultTimeout(20 * time.Minute),
		Default: schema.DefaultTimeout(20 * time.Minute),
	}
}
