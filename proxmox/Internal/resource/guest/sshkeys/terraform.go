package sshkeys

import (
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config []pveAPI.AuthorizedKey, d *schema.ResourceData) {
	keys := make([]string, len(config))
	for i := range config {
		keys[i] = config[i].String() + "\n"
	}
	d.Set(Root, strings.Join(keys, ""))
}
