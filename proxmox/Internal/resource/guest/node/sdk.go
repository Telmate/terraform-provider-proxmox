package node

import (
	"math/rand"
	"time"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SdkUpdate(d *schema.ResourceData, current pveAPI.NodeName) pveAPI.NodeName {
	if node, ok := d.GetOk(RootNode); ok {
		return pveAPI.NodeName(node.(string))
	}
	nodes := d.Get(RootNodes).(*schema.Set).List()
	currentNode := string(current)
	if len(nodes) == 1 {
		if currentNode != nodes[0].(string) {
			return pveAPI.NodeName(nodes[0].(string))
		}
		return current
	}
	if inArray(nodes, currentNode) {
		return current
	}
	randomIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))
	return pveAPI.NodeName(nodes[randomIndex].(string))
}

func SdkCreate(d *schema.ResourceData) pveAPI.NodeName {
	if node, ok := d.GetOk(RootNode); ok {
		return pveAPI.NodeName(node.(string))
	}
	nodes := d.Get(RootNodes).(*schema.Set).List()
	if len(nodes) == 1 {
		return pveAPI.NodeName(nodes[0].(string))
	}
	randomIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))
	return pveAPI.NodeName(nodes[randomIndex].(string))
}

func inArray(nodes []interface{}, current string) bool {
	for i := range nodes {
		if current == nodes[i].(string) {
			return true
		}
	}
	return false
}
