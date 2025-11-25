package proxmox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDriftDeletionDiagnostic(d *schema.ResourceData) diag.Diagnostic {
	id := d.Id()
	d.SetId("")
	return diag.Diagnostic{
		Summary:  "State drift detected: Resource (" + id + ") was deleted outside Terraform",
		Severity: diag.Warning}
}
