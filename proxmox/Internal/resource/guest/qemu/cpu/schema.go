package cpu

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root             string = "cpu"
	RootCores        string = "cores"
	RootCpuType      string = "cpu_type"
	RootNuma         string = "numa"
	RootSockets      string = "sockets"
	RootVirtualCores string = "vcpus"
        RootCpuAffinity  string = "cpu_affinity"
)

func SchemaCores() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  1,
		ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Errorf(RootCores + " must be an integer")
			}
			if v < 1 {
				return diag.Errorf(RootCores + " must be greater than 0")
			}
			return diag.FromErr(pveAPI.QemuCpuCores(v).Validate())
		}}
}

func SchemaType(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	return &s
}

func SchemaNuma() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
	}
}

func SchemaSockets() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  1,
		ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Errorf(RootSockets + " must be an integer")
			}
			if v < 1 {
				return diag.Errorf(RootSockets + " must be greater than 0")
			}
			return diag.FromErr(pveAPI.QemuCpuSockets(v).Validate())
		}}
}

func SchemaVirtualCores() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Errorf(RootVirtualCores + " must be an integer")
			}
			if v < 0 {
				return diag.Errorf(RootVirtualCores + " must be greater than or equal to 0")
			}
			return nil
		},
	}
}

func SchemaCpuAffinity() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
                Elem:          &schema.Schema{Type: schema.TypeInt},
        }
}
