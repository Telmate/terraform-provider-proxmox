package sshkeys

import (
	"regexp"
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root string = "sshkeys"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return trim(old) == trim(new)
		},
		ValidateDiagFunc: func(i interface{}, p cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(Root + " must be a string")
			}
			if v == "" {
				return nil
			}
			rawKeys := strings.Split(v, "\n")
			for i := range rawKeys {
				err := (&pveAPI.AuthorizedKey{}).Parse([]byte(rawKeys[i]))
				if err != nil {
					if strings.ReplaceAll(strings.ReplaceAll(rawKeys[i], "\t", ""), " ", "") == "" { // skip empty lines
						continue
					}
					return diag.Diagnostics{{
						Severity: diag.Error,
						Summary:  err.Error()}}
				}
			}
			return nil
		}}
}

var regexMultipleSpaces = regexp.MustCompile(`\s+`)

func trim(rawKeys string) string {
	return regexMultipleSpaces.ReplaceAllString(strings.TrimSpace(rawKeys), " ")
}
