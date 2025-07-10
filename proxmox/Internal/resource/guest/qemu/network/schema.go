package network

import (
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/mac"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/vlan/native"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root string = "network"

	AmountNetworkInterfaces int = int(pveAPI.QemuNetworkInterfacesAmount)

	schemaID string = "id"

	schemaBridge     string = "bridge"
	schemaFirewall   string = "firewall"
	schemaLinkDown   string = "link_down"
	schemaMAC        string = "macaddr"
	schemaMTU        string = "mtu"
	schemaModel      string = "model"
	schemaNativeVlan string = "tag"
	schemaQueues     string = "queues"
	schemaRate       string = "rate"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: AmountNetworkInterfaces,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeInt,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v < 0 {
							return diag.Errorf("%s must be in the range 0 - %d, got: %d", schemaID, pveAPI.QemuNetworkInterfaceIDMaximum, v)
						}
						return diag.FromErr(pveAPI.QemuNetworkInterfaceID(v).Validate())
					}},
				schemaBridge: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "nat"},
				schemaFirewall: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false},
				schemaLinkDown: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false},
				schemaMAC: mac.Schema(true, schemaMAC),
				schemaMTU: {
					Type:     schema.TypeInt,
					Optional: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v == 1 {
							return nil
						}
						if v < 0 {
							return diag.Errorf("%s must be equal or greater than 0, got: %d", schemaMTU, v)
						}
						if err := pveAPI.MTU(v).Validate(); err != nil {
							return diag.Errorf("%s must be in the range 576 - 65520, or 1 got: %d", schemaMTU, v)
						}
						return nil
					}},
				schemaModel: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: false,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(string)
						if err := pveAPI.QemuNetworkModel(v).Validate(); err != nil {
							return diag.Errorf("invalid network %s: %s", schemaModel, v)
						}
						return nil
					}},
				schemaNativeVlan: native.Schema(true, schemaNativeVlan),
				schemaQueues: {
					Type:     schema.TypeInt,
					Optional: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v < 0 {
							return diag.Errorf("%s must be equal or greater than 0, got: %d", schemaQueues, v)
						}
						if err := pveAPI.QemuNetworkQueue(v).Validate(); err != nil {
							return diag.Errorf("%s must be in the range 0 - %d, got: %d", schemaQueues, pveAPI.QemuNetworkQueueMaximum, v)
						}
						return nil
					}},
				schemaRate: {
					Type:     schema.TypeInt,
					Optional: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v < 0 {
							return diag.Errorf("%s must be equal or greater than 0, got: %d", schemaRate, v)
						}
						return nil
					}}}}}
}
