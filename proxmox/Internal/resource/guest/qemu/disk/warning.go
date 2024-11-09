package disk

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

func warningDisk(slot, setting, property, value, extra string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "slot: " + slot + " " + setting + " is ignored when " + property + " = " + value + extra}
}

func warningsCdromAndCloudinit(slot, kind string, schema map[string]interface{}) (diags diag.Diagnostics) {
	if schema["asyncio"].(string) != "" {
		diags = append(diags, warningDisk(slot, "asyncio", "type", kind, ""))
	}
	if schema["cache"].(string) != "" {
		diags = append(diags, warningDisk(slot, "cache", "type", kind, ""))
	}
	if schema["discard"].(bool) {
		diags = append(diags, warningDisk(slot, "discard", "type", kind, ""))
	}
	if schema["disk_file"].(string) != "" {
		diags = append(diags, warningDisk(slot, "disk_file", "type", kind, ""))
	}
	if schema["emulatessd"].(bool) {
		diags = append(diags, warningDisk(slot, "emulatessd", "type", kind, ""))
	}
	if schema["format"].(string) != "" {
		diags = append(diags, warningDisk(slot, "format", "type", kind, ""))
	}
	if schema["iops_r_burst"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_burst", "type", kind, ""))
	}
	if schema["iops_r_burst_length"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_burst_length", "type", kind, ""))
	}
	if schema["iops_r_concurrent"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_concurrent", "type", kind, ""))
	}
	if schema["iops_wr_burst"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_burst", "type", kind, ""))
	}
	if schema["iops_wr_burst_length"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_burst_length", "type", kind, ""))
	}
	if schema["iops_wr_concurrent"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_concurrent", "type", kind, ""))
	}
	if schema["iothread"].(bool) {
		diags = append(diags, warningDisk(slot, "iothread", "type", kind, ""))
	}
	if schema["mbps_r_burst"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_r_burst", "type", kind, ""))
	}
	if schema["mbps_r_concurrent"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_r_concurrent", "type", kind, ""))
	}
	if schema["mbps_wr_burst"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_wr_burst", "type", kind, ""))
	}
	if schema["mbps_wr_concurrent"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_wr_concurrent", "type", kind, ""))
	}
	if schema["readonly"].(bool) {
		diags = append(diags, warningDisk(slot, "readonly", "type", kind, ""))
	}
	if schema["replicate"].(bool) {
		diags = append(diags, warningDisk(slot, "replicate", "type", kind, ""))
	}
	if schema["serial"].(string) != "" {
		diags = append(diags, warningDisk(slot, "serial", "type", kind, ""))
	}
	if schema["size"].(string) != "" {
		diags = append(diags, warningDisk(slot, "size", "type", kind, ""))
	}
	if schema["wwn"].(string) != "" {
		diags = append(diags, warningDisk(slot, "wwn", "type", kind, ""))
	}
	return
}

func warningsDiskPassthrough(slot string, schema map[string]interface{}) diag.Diagnostics {
	if schema["format"].(string) != "" {
		return diag.Diagnostics{warningDisk(slot, "format", "type", "passthrough", "and slot = "+slot)}
	}
	if schema["storage"].(string) != "" {
		return diag.Diagnostics{warningDisk(slot, "storage", "type", "passthrough", "and slot = "+slot)}
	}
	return diag.Diagnostics{}
}
