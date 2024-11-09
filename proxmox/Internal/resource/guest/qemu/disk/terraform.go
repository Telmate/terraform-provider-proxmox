package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(d *schema.ResourceData, config pveAPI.QemuStorages) {
	if _, ok := d.GetOk("disk"); ok {
		d.Set("disk", terraform_Disk_QemuDisks(config))
	} else {
		d.Set("disks", terraform_Disks_QemuDisks(config))
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
	params["mbps_r_burst"] = float64(config.MBps.ReadLimit.Burst)
	params["mbps_r_concurrent"] = float64(config.MBps.ReadLimit.Concurrent)
	params["mbps_wr_burst"] = float64(config.MBps.WriteLimit.Burst)
	params["mbps_wr_concurrent"] = float64(config.MBps.WriteLimit.Concurrent)
	params["iops_r_burst"] = int(config.Iops.ReadLimit.Burst)
	params["iops_r_burst_length"] = int(config.Iops.ReadLimit.BurstDuration)
	params["iops_r_concurrent"] = int(config.Iops.ReadLimit.Concurrent)
	params["iops_wr_burst"] = int(config.Iops.WriteLimit.Burst)
	params["iops_wr_burst_length"] = int(config.Iops.WriteLimit.BurstDuration)
	params["iops_wr_concurrent"] = int(config.Iops.WriteLimit.Concurrent)
}
