package network

import (
	"net"

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
						err := pveAPI.QemuNetworkInterfaceID(v).Validate()
						if err != nil {
							return diag.Errorf("%s must be in the range 0 - %d, got: %d", schemaID, pveAPI.QemuNetworkInterfaceIDMaximum, v)
						}
						return nil
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
				schemaMAC: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						oldMAC, _ := net.ParseMAC(old)
						newMAC, _ := net.ParseMAC(new)
						return oldMAC.String() == newMAC.String()
					},
					ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
						v := i.(string)
						if _, err := net.ParseMAC(v); err != nil {
							return diag.Errorf("invalid %s: %s", schemaMAC, v)
						}
						return nil
					}},
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
				schemaNativeVlan: {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "VLAN tag.",
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v < 0 {
							return diag.Errorf("%s must be equal or greater than 0, got: %d", schemaNativeVlan, v)
						}
						return nil
					}},
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
