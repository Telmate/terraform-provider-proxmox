package disk

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

func warningDisk(slot, setting, property, value, extra string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  schemaSlot + ": " + slot + " " + setting + " is ignored when " + property + " = " + value + extra}
}

func warningsCdromAndCloudinit(slot, kind string, schema map[string]interface{}) (diags diag.Diagnostics) {
	if schema[schemaAsyncIO].(string) != "" {
		diags = diag.Diagnostics{warningDisk(slot, schemaAsyncIO, schemaType, kind, "")}
	}
	if schema[schemaCache].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaCache, schemaType, kind, ""))
	}
	if schema[schemaDiscard].(bool) {
		diags = append(diags, warningDisk(slot, schemaDiscard, schemaType, kind, ""))
	}
	if schema[schemaDiskFile].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, kind, ""))
	}
	if schema[schemaEmulateSSD].(bool) {
		diags = append(diags, warningDisk(slot, schemaEmulateSSD, schemaType, kind, ""))
	}
	if schema[schemaFormat].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaFormat, schemaType, kind, ""))
	}
	if schema[schemaIOPSrBurst].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrBurst, schemaType, kind, ""))
	}
	if schema[schemaIOPSrBurstLength].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrBurstLength, schemaType, kind, ""))
	}
	if schema[schemaIOPSrConcurrent].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrConcurrent, schemaType, kind, ""))
	}
	if schema[schemaIOPSwrBurst].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrBurst, schemaType, kind, ""))
	}
	if schema[schemaIOPSwrBurstLength].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrBurstLength, schemaType, kind, ""))
	}
	if schema[schemaIOPSwrConcurrent].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrConcurrent, schemaType, kind, ""))
	}
	if schema[schemaIOthread].(bool) {
		diags = append(diags, warningDisk(slot, schemaIOthread, schemaType, kind, ""))
	}
	if schema[schemaMBPSrBurst].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSrBurst, schemaType, kind, ""))
	}
	if schema[schemaMBPSrConcurrent].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSrConcurrent, schemaType, kind, ""))
	}
	if schema[schemaMBPSwrBurst].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSwrBurst, schemaType, kind, ""))
	}
	if schema[schemaMBPSwrConcurrent].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSwrConcurrent, schemaType, kind, ""))
	}
	if schema[schemaReadOnly].(bool) {
		diags = append(diags, warningDisk(slot, schemaReadOnly, schemaType, kind, ""))
	}
	if schema[schemaReplicate].(bool) {
		diags = append(diags, warningDisk(slot, schemaReplicate, schemaType, kind, ""))
	}
	if schema[schemaSerial].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaSerial, schemaType, kind, ""))
	}
	if schema[schemaSize].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaSize, schemaType, kind, ""))
	}
	if schema[schemaWorldWideName].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaWorldWideName, schemaType, kind, ""))
	}
	return
}

func warningsDiskPassthrough(slot string, schema map[string]interface{}) diag.Diagnostics {
	if schema[schemaFormat].(string) != "" {
		return diag.Diagnostics{warningDisk(slot, schemaFormat, schemaType, schemaPassthrough, "and "+schemaSlot+" = "+slot)}
	}
	if schema[schemaStorage].(string) != "" {
		return diag.Diagnostics{warningDisk(slot, schemaStorage, schemaType, schemaPassthrough, "and "+schemaSlot+" = "+slot)}
	}
	return diag.Diagnostics{}
}

func WarningIgnore(slot string, schema map[string]any) (diags diag.Diagnostics) {
	if schema[schemaAsyncIO].(string) != "" {
		diags = diag.Diagnostics{warningDisk(slot, schemaAsyncIO, schemaType, enumIgnore, "")}
	}
	if !schema[schemaBackup].(bool) {
		diags = append(diags, warningDisk(slot, schemaBackup, schemaType, enumIgnore, ""))
	}
	if schema[schemaCache].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaCache, schemaType, enumIgnore, ""))
	}
	if schema[schemaDiscard].(bool) {
		diags = append(diags, warningDisk(slot, schemaDiscard, schemaType, enumIgnore, ""))
	}
	if schema[schemaDiskFile].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, enumIgnore, ""))
	}
	if schema[schemaEmulateSSD].(bool) {
		diags = append(diags, warningDisk(slot, schemaEmulateSSD, schemaType, enumIgnore, ""))
	}
	if schema[schemaFormat].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaFormat, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSrBurst].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrBurst, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSrBurstLength].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrBurstLength, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSrConcurrent].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSrConcurrent, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSwrBurst].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrBurst, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSwrBurstLength].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrBurstLength, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOPSwrConcurrent].(int) != 0 {
		diags = append(diags, warningDisk(slot, schemaIOPSwrConcurrent, schemaType, enumIgnore, ""))
	}
	if schema[schemaIOthread].(bool) {
		diags = append(diags, warningDisk(slot, schemaIOthread, schemaType, enumIgnore, ""))
	}
	if schema[schemaISO].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaISO, schemaType, enumIgnore, ""))
	}
	if schema[schemaMBPSrBurst].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSrBurst, schemaType, enumIgnore, ""))
	}
	if schema[schemaMBPSrConcurrent].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSrConcurrent, schemaType, enumIgnore, ""))
	}
	if schema[schemaMBPSwrBurst].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSwrBurst, schemaType, enumIgnore, ""))
	}
	if schema[schemaMBPSwrConcurrent].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, schemaMBPSwrConcurrent, schemaType, enumIgnore, ""))
	}
	if schema[schemaPassthrough].(bool) {
		diags = append(diags, warningDisk(slot, schemaPassthrough, schemaType, enumIgnore, ""))
	}
	if schema[schemaReadOnly].(bool) {
		diags = append(diags, warningDisk(slot, schemaReadOnly, schemaType, enumIgnore, ""))
	}
	if schema[schemaReplicate].(bool) {
		diags = append(diags, warningDisk(slot, schemaReplicate, schemaType, enumIgnore, ""))
	}
	if schema[schemaSerial].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaSerial, schemaType, enumIgnore, ""))
	}
	if schema[schemaSize].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaSize, schemaType, enumIgnore, ""))
	}
	if schema[schemaStorage].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaStorage, schemaType, enumIgnore, ""))
	}
	if schema[schemaWorldWideName].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaWorldWideName, schemaType, enumIgnore, ""))
	}
	return
}
