package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(d *schema.ResourceData, config pveAPI.QemuStorages, ciDisk *bool) {
	if _, ok := d.GetOk(RootDisk); ok {
		d.Set(RootDisk, terraform_Disk_QemuDisks(config, ciDisk))
	} else {
		d.Set(RootDisks, terraform_Disks_QemuDisks(config, ciDisk))
	}
}

func terraformLinkedCloneId(id *uint) int {
	if id != nil {
		return int(*id)
	}
	return -1
}

func terraformIsoFile(config *pveAPI.IsoFile) string {
	if config == nil {
		return ""
	}
	return config.Storage + ":iso/" + config.File
}

func terraformQemuDiskBandwidth(params map[string]interface{}, config pveAPI.QemuDiskBandwidth) {
	params[schemaMBPSrBurst] = float64(config.MBps.ReadLimit.Burst)
	params[schemaMBPSrConcurrent] = float64(config.MBps.ReadLimit.Concurrent)
	params[schemaMBPSwrBurst] = float64(config.MBps.WriteLimit.Burst)
	params[schemaMBPSwrConcurrent] = float64(config.MBps.WriteLimit.Concurrent)
	params[schemaIOPSrBurst] = int(config.Iops.ReadLimit.Burst)
	params[schemaIOPSrBurstLength] = int(config.Iops.ReadLimit.BurstDuration)
	params[schemaIOPSrConcurrent] = int(config.Iops.ReadLimit.Concurrent)
	params[schemaIOPSwrBurst] = int(config.Iops.WriteLimit.Burst)
	params[schemaIOPSwrBurstLength] = int(config.Iops.WriteLimit.BurstDuration)
	params[schemaIOPSwrConcurrent] = int(config.Iops.WriteLimit.Concurrent)
}
