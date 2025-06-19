package errorMSG

import (
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
