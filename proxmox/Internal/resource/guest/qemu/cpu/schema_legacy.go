package cpu

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootLegacyCores        = "cores"
	RootLegacyCpuType      = "cpu_type"
	RootLegacyNuma         = "numa"
	RootLegacySockets      = "sockets"
	RootLegacyVirtualCores = "vcpus"
)

func SchemaLegacyCores() *schema.Schema {
	return subSchemaCores(RootLegacyCores, schema.Schema{
		Deprecated:    "use '" + Root + " { " + RootLegacyCores + " = }' instead",
		ConflictsWith: []string{Root}})
}

func SchemaLegacyType() *schema.Schema {
	return subSchemaType(schema.Schema{
		Deprecated:    "use '" + Root + " { " + RootLegacyCpuType + " = }' instead",
		ConflictsWith: []string{Root},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			newNew := defaultType
			if new != "" {
				newNew = new
			}
			return old == newNew
		}})
}

func SchemaLegacyNuma() *schema.Schema {
	return subSchemaNuma(schema.Schema{
		Deprecated:    "use '" + Root + " { " + RootLegacyNuma + " = }' instead",
		ConflictsWith: []string{Root}})
}

func SchemaLegacySockets() *schema.Schema {
	return subSchemaSockets(RootLegacySockets, schema.Schema{
		Deprecated:    "use '" + Root + " { " + RootLegacySockets + " = }' instead",
		ConflictsWith: []string{Root},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			newNew := strconv.Itoa(defaultSockets)
			if new != "0" {
				newNew = new
			}
			return old == newNew
		}})
}

func SchemaLegacyVirtualCores() *schema.Schema {
	return subSchemaVirtualCores(RootLegacyVirtualCores, schema.Schema{
		Deprecated:    "use '" + Root + " { " + RootLegacyVirtualCores + " = }' instead",
		ConflictsWith: []string{Root}})
}
