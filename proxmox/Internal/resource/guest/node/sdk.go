package node

import (
	"errors"
	"math/rand"
	"time"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const errorNoNodeConfigured = "no target node specified"

func SdkUpdate(d *schema.ResourceData, current pveAPI.NodeName) (pveAPI.NodeName, error) {
	if node, ok := d.GetOk(RootNode); ok {
		return pveAPI.NodeName(node.(string)), nil
	}
	nodes := d.Get(RootNodes).(*schema.Set).List()
	currentNode := string(current)
	switch len(nodes) {
	case 0:
		return "", errors.New(errorNoNodeConfigured)
	case 1:
		if currentNode != nodes[0].(string) {
			return pveAPI.NodeName(nodes[0].(string)), nil
		}
		return current, nil
	}
	if inArray(nodes, currentNode) {
		return current, nil
	}
	randomIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))
	return pveAPI.NodeName(nodes[randomIndex].(string)), nil
}

// SdkCreate selects a node for resource creation.
func SdkCreate(d *schema.ResourceData) (pveAPI.NodeName, error) {
	if node, ok := d.GetOk(RootNode); ok {
		return pveAPI.NodeName(node.(string)), nil
	}
	nodes := d.Get(RootNodes).(*schema.Set).List()
	switch len(nodes) {
	case 0:
		return "", errors.New(errorNoNodeConfigured)
	case 1:
		return pveAPI.NodeName(nodes[0].(string)), nil
	}
	randomIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))
	return pveAPI.NodeName(nodes[randomIndex].(string)), nil
}

func inArray(nodes []any, current string) bool {
	for i := range nodes {
		if current == nodes[i].(string) {
			return true
		}
	}
	return false
}
