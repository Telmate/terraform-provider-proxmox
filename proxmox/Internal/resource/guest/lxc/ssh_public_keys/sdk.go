package ssh_public_keys

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) []pveSDK.AuthorizedKey {
	v := d.Get(Root)
	rawKeys := strings.Split(v.(string), "\n")
	keys := make([]pveSDK.AuthorizedKey, len(rawKeys))
	for i := range rawKeys {
		tmpKey := &pveSDK.AuthorizedKey{}
		_ = tmpKey.Parse([]byte(rawKeys[i]))
		keys[i] = *tmpKey
	}
	return keys
}
