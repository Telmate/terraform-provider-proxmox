package boot

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveSDK.BootOrder, d *schema.ResourceData) {
	old := d.Get(Root).(string)
	if equal(toBoot(old), config) {
		d.Set(Root, old)
		return
	}
	var b strings.Builder
	for i := range config {
		b.WriteString(config[i].String())
		b.WriteByte(separator)
	}
	var out string
	if b.Len() > 0 {
		out = b.String()[0 : b.Len()-1]
	}
	d.Set(Root, out)
}

func equal(a, b pveSDK.BootOrder) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
