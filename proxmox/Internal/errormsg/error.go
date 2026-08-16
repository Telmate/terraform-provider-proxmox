package errorMSG

import (
	"errors"
	"strings"

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

type PathPart struct {
	Name  string
	Array bool
}

func (p PathPart) close() byte {
	if p.Array {
		return ']'
	}
	return '}'
}

func (p PathPart) open() byte {
	if p.Array {
		return '['
	}
	return '{'
}

type TfProperty struct {
	Path  []PathPart
	Value *string
}

func (p TfProperty) build(b *strings.Builder) {
	for i := range p.Path {
		b.WriteString(p.Path[i].Name)
		if i+1 == len(p.Path) {
			if p.Value != nil {
				b.WriteByte('=')
				b.WriteString(*p.Value)
			}
			break
		}
		b.WriteByte(p.Path[i].open())
	}
	if len(p.Path) > 1 {
		for i := range len(p.Path) - 1 {
			b.WriteByte(p.Path[i].close())
		}
	}
}

const (
	incompatibleConfig = "Incompatible configuration\n\n  "
	isIncompatible     = ": is incompatible with "
)

func ConflictsWith(a, b TfProperty) error {
	var builder strings.Builder
	builderPTR := &builder
	builder.WriteString(incompatibleConfig)
	a.build(builderPTR)
	builder.WriteString(isIncompatible)
	b.build(builderPTR)
	return errors.New(builder.String())
}

func ConflictsWithWhere(a, b, where TfProperty) error {
	var builder strings.Builder
	builderPTR := &builder
	builder.WriteString(incompatibleConfig)
	a.build(builderPTR)
	builder.WriteString(isIncompatible)
	b.build(builderPTR)
	builder.WriteString("\n  where ")
	where.build(builderPTR)
	return errors.New(builder.String())
}
