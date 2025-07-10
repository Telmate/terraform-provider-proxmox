package errorMSG

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	Uint   string = "expected type of %s to be a positive number (uint)"
	Float  string = "expected type of %s to be a float"
	String string = "expected type of %s to be string"
)

func UintDiagnostic(k string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Invalid type",
		Detail:   "expected type of " + k + " to be a positive number (uint)"}
}

func UintDiagnostics(k string) diag.Diagnostics {
	return diag.Diagnostics{UintDiagnostic(k)}
}

func StringDiagnostic(k string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Invalid type",
		Detail:   "expected type of " + k + " to be a string"}
}

func StringDiagnostics(k string) diag.Diagnostics {
	return diag.Diagnostics{StringDiagnostic(k)}
}

type Diagnostic struct {
	Severity         diag.Severity
	Summary          string
	Detail           string
	AttributePath    cty.Path
	UseAttributePath bool
}

func (d Diagnostic) Diagnostic() diag.Diagnostic {
	var k cty.Path
	if d.UseAttributePath {
		k = d.AttributePath
	}
	if d.Summary == "" {
		d.Summary = d.Detail
	}
	return diag.Diagnostic{
		AttributePath: k,
		Detail:        d.Detail,
		Severity:      d.Severity,
		Summary:       d.Summary}
}

func (d Diagnostic) Diagnostics() diag.Diagnostics {
	return diag.Diagnostics{d.Diagnostic()}
}
