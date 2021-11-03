package proxmox

import (
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTimeouts() *schema.ResourceTimeout {
	resourceCreateTimeout := defaultTimeout
	resourceReadTimeout := 180
	resourceUpdateTimeout := 180
	resourceDeleteTimeout := 180

	if v, ok := os.LookupEnv("PM_TIMEOUT"); ok {
		resourceCreateTimeout, _ = strconv.Atoi(v)
	}

	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(time.Duration(resourceCreateTimeout) * time.Second),
		Read:    schema.DefaultTimeout(time.Duration(resourceReadTimeout) * time.Second),
		Update:  schema.DefaultTimeout(time.Duration(resourceUpdateTimeout) * time.Second),
		Delete:  schema.DefaultTimeout(time.Duration(resourceDeleteTimeout) * time.Second),
		Default: schema.DefaultTimeout(defaultTimeout * time.Second),
	}
}
