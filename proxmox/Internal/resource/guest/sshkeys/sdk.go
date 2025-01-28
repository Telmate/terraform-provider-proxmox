package sshkeys

import (
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *[]pveAPI.AuthorizedKey {
	v := d.Get(Root)
	rawKeys := strings.Split(v.(string), "\n")
	keys := make([]pveAPI.AuthorizedKey, len(rawKeys))
	for i := range rawKeys {
		tmpKey := &pveAPI.AuthorizedKey{}
		_ = tmpKey.Parse([]byte(rawKeys[i]))
		keys[i] = *tmpKey
	}
	return &keys
}
