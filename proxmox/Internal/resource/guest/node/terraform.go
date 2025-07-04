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
	nodes := d.Get(RootNodes).(*schema.Set).List()
	if inArray(nodes, current) {
		d.Set(RootNodes, nodes)
		return
	}
	d.Set(RootNodes, []interface{}{current})
}
