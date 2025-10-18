package networks

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/mac"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/vlan/native"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootNetwork  = "network"
	RootNetworks = "networks"

	prefixSchemaID = "net"

	schemaID = "id"

	schemaBridge     = "bridge"
	schemaConnected  = "connected"
	schemaFirewall   = "firewall"
	schemaMAC        = mac.Root
	schemaMTU        = "mtu"
	schemaName       = "name"
	schemaNativeVlan = native.Root
	schemaRateLimit  = "rate_limit"

	schmemaIPv4 = "ipv4"
	schmemaIPv6 = "ipv6"

	schemaAddress = "address"
	schemaDHCP    = "dhcp"
	schemaGateway = "gateway"
	schemaSLAAC   = "slaac"

	schemaIPv4Address = "ipv4_" + schemaAddress
	schemaIPv4DHCP    = "ipv4_" + schemaDHCP
	schemaIPv4Gateway = "ipv4_" + schemaGateway
	schemaIPv6Address = "ipv6_" + schemaAddress
	schemaIPv6DHCP    = "ipv6_" + schemaDHCP
	schemaIPv6Gateway = "ipv6_" + schemaGateway

	networksAmount = pveSDK.LxcNetworksAmount
	maximumID      = pveSDK.LxcNetworkIdMaximum

	defaultConnected = true
	defaultFirewall  = true
)

func subSchemaBridge() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true}
}

func subSchemaConnected() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  defaultConnected}
}

func subSchemaFirewall() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  defaultFirewall}
}

func subSchemaMAC(useAttributePath bool, path string) *schema.Schema {
	return mac.Schema(useAttributePath, path)
}

func subSchemaMTU(useAttributePath bool, path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
			v := i.(int)
			if err := pveSDK.MTU(v).Validate(); err != nil {
				return errorMSG.Diagnostic{
					Summary:          "invalid " + path,
					Severity:         diag.Error,
					UseAttributePath: useAttributePath,
					AttributePath:    k}.Diagnostics()
			}
			return nil
		}}
}

func subSchemaName(useAttributePath bool, path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
			v := i.(string)
			if pveSDK.LxcNetworkName(v).Validate() != nil {
				return errorMSG.Diagnostic{
					Summary:          "invalid " + path + ": " + v,
					Severity:         diag.Error,
					UseAttributePath: useAttributePath,
					AttributePath:    k}.Diagnostics()
			}
			return nil
		}}
}

func subSchemaNativeVlan(useAttributePath bool, path string) *schema.Schema {
	return native.Schema(useAttributePath, path)
}

func subSchemaRate(useAttributePath bool, path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v := i.(int)
			if v < 0 {
				return errorMSG.Diagnostic{
					Severity:         diag.Error,
					Summary:          path + " must be equal or greater than 0, got: " + strconv.Itoa(v),
					UseAttributePath: useAttributePath,
					AttributePath:    k}.Diagnostics()
			}
			return nil
		}}
}

func subSchemaDHCP(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeBool
	s.Optional = true
	return &s
}

func subSchemaSLAAC(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeBool
	s.Optional = true
	return &s
}

func subSchemaIPv4Address(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v := i.(string)
		if v == "" {
			return nil
		}
		if err := pveSDK.IPv4CIDR(v).Validate(); err != nil {
			return errorMSG.Diagnostic{
				Summary:          "invalid " + path + ": " + v,
				Severity:         diag.Error,
				UseAttributePath: useAttributePath,
				AttributePath:    k}.Diagnostics()
		}
		return nil
	}
	return &s
}

func subSchemaIPv4Gateway(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v := i.(string)
		if v == "" {
			return nil
		}
		if err := pveSDK.IPv4Address(v).Validate(); err != nil {
			return errorMSG.Diagnostic{
				Summary:          "invalid " + path + ": " + v,
				Severity:         diag.Error,
				UseAttributePath: useAttributePath,
				AttributePath:    k}.Diagnostics()
		}
		return nil
	}
	return &s
}

func subSchemaIPv6Address(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v := i.(string)
		if v == "" {
			return nil
		}
		if err := pveSDK.IPv6CIDR(v).Validate(); err != nil {
			return errorMSG.Diagnostic{
				Summary:          "invalid " + path + ": " + v,
				Severity:         diag.Error,
				UseAttributePath: useAttributePath,
				AttributePath:    k}.Diagnostics()
		}
		return nil
	}
	return &s
}

func subSchemaIPv6Gateway(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v := i.(string)
		if v == "" {
			return nil
		}
		if err := pveSDK.IPv6Address(v).Validate(); err != nil {
			return errorMSG.Diagnostic{
				Summary:          "invalid " + path + ": " + v,
				Severity:         diag.Error,
				UseAttributePath: useAttributePath,
				AttributePath:    k}.Diagnostics()
		}
		return nil
	}
	return &s
}
