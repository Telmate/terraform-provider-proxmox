// Package rng provides the random number generator device.
package rng

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "rng"

	schemaLimit  = "limit"
	schemaPeriod = "period"
	schemaSource = "source"

	uRandom = pveSDK.EntropySourceRawURandom
	random  = pveSDK.EntropySourceRawRandom
	hwRNG   = pveSDK.EntropySourceRawHwRNG

	defaultLimit  = 1024
	defaultSource = uRandom
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaLimit:  subSchemaLimit(),
				schemaPeriod: subSchemaPeriod(),
				schemaSource: subSchemaSource()}}}
}

func subSchemaLimit() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  defaultLimit,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			if v, ok := i.(int); ok && v >= 0 {
				return nil
			}
			return errorMSG.UintDiagnostics(schemaLimit)
		}}
}

func subSchemaPeriod() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			if v, ok := i.(int); ok && v >= 0 {
				return nil
			}
			return errorMSG.UintDiagnostics(schemaPeriod)
		}}
}

func subSchemaSource() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  defaultSource,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			if v, ok := i.(string); ok {
				switch v {
				case uRandom, random, hwRNG:
					return nil
				}
			}
			return errorMSG.Diagnostic{
				Summary:  "expected type of " + schemaSource + " to be one of " + uRandom + ", " + random + ", or " + hwRNG,
				Severity: diag.Error}.Diagnostics()
		}}
}
