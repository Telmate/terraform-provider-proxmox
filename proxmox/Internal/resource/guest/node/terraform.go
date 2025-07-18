package node

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(currentNode pveAPI.NodeName, d *schema.ResourceData) {
	current := string(currentNode)
	d.Set(Computed, current)
	if _, ok := d.GetOk(RootNode); ok {
		d.Set(RootNode, current)
		return
	}
	if v, ok := d.GetOk(RootNodes); ok && v != nil {
		nodes := v.(*schema.Set).List()
		if inArray(nodes, current) {
			d.Set(RootNodes, nodes)
			return
		}
	}
	d.Set(RootNodes, []any{current})
}
