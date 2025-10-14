package clone

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Settings struct {
	ID   *pveSDK.GuestID
	Name *pveSDK.GuestName
	Node pveSDK.NodeName
	Pool *pveSDK.PoolName
}

type Return struct {
	ID     pveSDK.GuestID
	Name   pveSDK.GuestName
	Target pveSDK.CloneLxcTarget
}

func SDK(d *schema.ResourceData, s Settings) *Return {
	v, ok := d.GetOk(Root)
	if !ok {
		return nil
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return nil
	}
	settings, ok := vv[0].(map[string]any)
	if !ok {
		return nil
	}
	if settings[schemaLinked].(bool) {
		return &Return{
			ID:   pveSDK.GuestID(settings[SchemaID].(int)),
			Name: pveSDK.GuestName(settings[SchemaName].(string)),
			Target: pveSDK.CloneLxcTarget{Linked: &pveSDK.CloneLinked{
				ID:   s.ID,
				Name: s.Name,
				Node: s.Node}}}
	}
	return &Return{
		ID:   pveSDK.GuestID(settings[SchemaID].(int)),
		Name: pveSDK.GuestName(settings[SchemaName].(string)),
		Target: pveSDK.CloneLxcTarget{Full: &pveSDK.CloneLxcFull{
			ID:   s.ID,
			Name: s.Name,
			Node: s.Node,
			Pool: s.Pool}}}
}
