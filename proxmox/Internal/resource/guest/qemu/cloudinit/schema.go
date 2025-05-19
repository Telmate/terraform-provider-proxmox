package cloudinit

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootCustom       = "cicustom"
	RootNameServers  = "nameserver"
	RootPassword     = "cipassword"
	RootSearchDomain = "searchdomain"
	RootUpgrade      = "ciupgrade"
	RootUser         = "ciuser"

	RootNetworkConfig0  = prefixNetworkConfig + "0"
	RootNetworkConfig1  = prefixNetworkConfig + "1"
	RootNetworkConfig2  = prefixNetworkConfig + "2"
	RootNetworkConfig3  = prefixNetworkConfig + "3"
	RootNetworkConfig4  = prefixNetworkConfig + "4"
	RootNetworkConfig5  = prefixNetworkConfig + "5"
	RootNetworkConfig6  = prefixNetworkConfig + "6"
	RootNetworkConfig7  = prefixNetworkConfig + "7"
	RootNetworkConfig8  = prefixNetworkConfig + "8"
	RootNetworkConfig9  = prefixNetworkConfig + "9"
	RootNetworkConfig10 = prefixNetworkConfig + "10"
	RootNetworkConfig11 = prefixNetworkConfig + "11"
	RootNetworkConfig12 = prefixNetworkConfig + "12"
	RootNetworkConfig13 = prefixNetworkConfig + "13"
	RootNetworkConfig14 = prefixNetworkConfig + "14"
	RootNetworkConfig15 = prefixNetworkConfig + "15"

	prefixNetworkConfig = "ipconfig"
)

func SchemaCiCustom() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		ForceNew: true,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return sdkCloudInitCustom(new).String() == sdkCloudInitCustom(old).String()
		}}
}

func SchemaNameServers() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return trimNameServers(old) == trimNameServers(new)
		}}
}

func SchemaPassword() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Optional:  true,
		Sensitive: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == "**********"
		}}
}

func SchemaSearchDomain() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true}
}

func SchemaUpgrade() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false}
}

func SchemaUser() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true}
}

func SchemaNetworkConfig() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true}
}
