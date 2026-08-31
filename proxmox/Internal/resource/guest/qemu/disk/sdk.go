package disk

import (
	"github.com/hashicorp/go-cty/cty"
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) (*pveAPI.QemuStorages, diag.Diagnostics) {
	if v, ok := d.GetOk(RootDisk); ok {
		diags := make(diag.Diagnostics, 0)
		storages := &pveAPI.QemuStorages{
			Ide:    sdk_Disks_QemuIdeDisksDefault(),
			Sata:   sdk_Disks_QemuSataDisksDefault(),
			Scsi:   sdk_Disks_QemuScsiDisksDefault(),
			VirtIO: sdk_Disks_QemuVirtIODisksDefault()}
		mapWriteOnlyImportFromDiskToSDK(d, v.([]interface{}))
		for _, disk := range v.([]interface{}) {
			tmpDisk := disk.(map[string]interface{})
			slot := tmpDisk[schemaSlot].(string)
			if len(slot) > 6 { // virtio
				diags = append(diags, sdk_Disk_QemuVirtIODisks(storages.VirtIO, slot[6:], tmpDisk)...)
				continue
			}
			if len(slot) > 4 {
				switch slot[0:4] {
				case schemaSata:
					diags = append(diags, sdk_Disk_QemuSataDisks(storages.Sata, slot[4:], tmpDisk)...)
				case schemaScsi:
					diags = append(diags, sdk_Disk_QemuScsiDisks(storages.Scsi, slot[4:], tmpDisk)...)
				}
				continue
			}
			if len(slot) > 3 { // ide
				diags = append(diags, sdk_Disk_QemuIdeDisks(storages.Ide, slot[3:], tmpDisk)...)
			}
		}
		return storages, diags
	} else if v, ok := d.GetOk(RootDisks); ok {
		if vv, ok := v.([]any); ok && len(vv) == 1 && vv[0] != nil {
			schemaStorages := vv[0].(map[string]any)
			mapWriteOnlyImportfromDisksToSDK(d, schemaStorages)
			return &pveAPI.QemuStorages{
				Ide:    sdk_Disks_QemuIdeDisks(schemaStorages),
				Sata:   sdk_Disks_QemuSataDisks(schemaStorages),
				Scsi:   sdk_Disks_QemuScsiDisks(schemaStorages),
				VirtIO: sdk_Disks_QemuVirtIODisks(schemaStorages)}, nil
		}
	}
	return &pveAPI.QemuStorages{
		Ide:    sdk_Disks_QemuIdeDisksDefault(),
		Sata:   sdk_Disks_QemuSataDisksDefault(),
		Scsi:   sdk_Disks_QemuScsiDisksDefault(),
		VirtIO: sdk_Disks_QemuVirtIODisksDefault()}, nil
}

func sdkIsoFile(iso string) *pveAPI.IsoFile {
	if iso == "" {
		return nil
	}
	storage, fileWithPrefix, cut := strings.Cut(iso, ":")
	if !cut {
		return nil
	}
	_, file, cut := strings.Cut(fileWithPrefix, "/")
	if !cut {
		return nil
	}
	return &pveAPI.IsoFile{File: file, Storage: storage}
}

func mapWriteOnlyImportFromDiskToSDK(d *schema.ResourceData, v []interface{}) {
	raw, _ := d.GetRawConfigAt(cty.GetAttrPath(RootDisk))
	if raw.CanIterateElements() {
		for _, rawDisk := range raw.AsValueSlice() {
			slot := rawDisk.GetAttr(schemaSlot).AsString()
			for _, vv := range v {
				if diskMap, ok := vv.(map[string]interface{}); ok && diskMap[schemaSlot] == slot {
					diskMap[schemaImportFrom] = rawDisk.GetAttr(schemaImportFrom).AsString()
				}
			}
		}
	}
}

func mapWriteOnlyImportfromDisksToSDK(d *schema.ResourceData, v map[string]any) {
	raw, _ := d.GetRawConfigAt(cty.GetAttrPath(RootDisks))
	if ide, ok := v[schemaIDE].([]interface{}); ok && len(ide) > 0 {
		mapWriteOnlyImportFromDisksToType(raw, ide, schemaIDE)
	}
	if sata, ok := v[schemaSata].([]interface{}); ok && len(sata) > 0 {
		mapWriteOnlyImportFromDisksToType(raw, sata, schemaSata)
	}
	if scsi, ok := v[schemaScsi].([]interface{}); ok && len(scsi) > 0 {
		mapWriteOnlyImportFromDisksToType(raw, scsi, schemaScsi)
	}
	if virtio, ok := v[schemaVirtIO].([]interface{}); ok && len(virtio) > 0 {
		mapWriteOnlyImportFromDisksToType(raw, virtio, schemaVirtIO)
	}
}

func mapWriteOnlyImportFromDisksToType(raw cty.Value, diskTypeData []interface{}, diskType string) {
	for key, typeSlot := range diskTypeData[0].(map[string]interface{}) {
		if slot, ok := typeSlot.([]interface{}); ok && len(slot) > 0 {
			if typeSlotDisks, ok := slot[0].(map[string]interface{}); ok {
				if typeSlotDisk, ok := typeSlotDisks[schemaDisk].([]interface{}); ok && len(typeSlotDisk) > 0 {
					tmpDisk := typeSlotDisk[0].(map[string]interface{})
					tmpDisk[schemaImportFrom] = getNestedImportFromRawRootDisks(raw, diskType, key)
				}
			}
		}
	}
}

func getNestedImportFromRawRootDisks(raw cty.Value, diskType string, diskSlot string) string {
	t := getFirstFromAttrList(raw.AsValueSlice()[0], diskType)
	s := getFirstFromAttrList(t, diskSlot)
	d := getFirstFromAttrList(s, schemaDisk)
	return d.GetAttr(schemaImportFrom).AsString()
}

func getFirstFromAttrList(raw cty.Value, key string) cty.Value {
	if attr := raw.GetAttr(key); !attr.IsNull() {
		if attr.CanIterateElements() && attr.LengthInt() > 0 {
			return attr.AsValueSlice()[0]
		}
	}
	return cty.NilVal
}
