package disk

import (
	"regexp"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootDisk  string = "disk"
	RootDisks string = "disks"
)

func SchemaDisk() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{RootDisks},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"asyncio":              subSchemaDiskAsyncIO(),
				"backup":               subSchemaDiskBackup(),
				"cache":                subSchemaDiskCache(),
				"discard":              {Type: schema.TypeBool, Optional: true},
				"disk_file":            subSchemaPassthroughFile(schema.Schema{Optional: true}),
				"emulatessd":           {Type: schema.TypeBool, Optional: true},
				"format":               subSchemaDiskFormat(schema.Schema{}),
				"id":                   subSchemaDiskId(),
				"iops_r_burst":         subSchemaDiskBandwidthIopsBurst(),
				"iops_r_burst_length":  subSchemaDiskBandwidthIopsBurstLength(),
				"iops_r_concurrent":    subSchemaDiskBandwidthIopsConcurrent(),
				"iops_wr_burst":        subSchemaDiskBandwidthIopsBurst(),
				"iops_wr_burst_length": subSchemaDiskBandwidthIopsBurstLength(),
				"iops_wr_concurrent":   subSchemaDiskBandwidthIopsConcurrent(),
				"iothread":             {Type: schema.TypeBool, Optional: true},
				"iso":                  subSchemaIsoPath(schema.Schema{}),
				"linked_disk_id":       subSchemaLinkedDiskId(),
				"mbps_r_burst":         subSchemaDiskBandwidthMBpsBurst(),
				"mbps_r_concurrent":    subSchemaDiskBandwidthMBpsConcurrent(),
				"mbps_wr_burst":        subSchemaDiskBandwidthMBpsBurst(),
				"mbps_wr_concurrent":   subSchemaDiskBandwidthMBpsConcurrent(),
				"passthrough":          {Type: schema.TypeBool, Default: false, Optional: true},
				"readonly":             {Type: schema.TypeBool, Optional: true},
				"replicate":            {Type: schema.TypeBool, Optional: true},
				"serial":               subSchemaDiskSerial(),
				"size":                 subSchemaDiskSize(schema.Schema{Computed: true, Optional: true}),
				"slot": {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Errorf(errorMSG.String, k)
						}
						switch v {
						case "ide0", "ide1", "ide2",
							"sata0", "sata1", "sata2", "sata3", "sata4", "sata5",
							"scsi0", "scsi1", "scsi2", "scsi3", "scsi4", "scsi5", "scsi6", "scsi7", "scsi8", "scsi9", "scsi10", "scsi11", "scsi12", "scsi13", "scsi14", "scsi15", "scsi16", "scsi17", "scsi18", "scsi19", "scsi20", "scsi21", "scsi22", "scsi23", "scsi24", "scsi25", "scsi26", "scsi27", "scsi28", "scsi29", "scsi30",
							"virtio0", "virtio1", "virtio2", "virtio3", "virtio4", "virtio5", "virtio6", "virtio7", "virtio8", "virtio9", "virtio10", "virtio11", "virtio12", "virtio13", "virtio14", "virtio15":
							return nil
						}
						return diag.Errorf("slot must be one of 'ide0', 'ide1', 'ide2', 'sata0', 'sata1', 'sata2', 'sata3', 'sata4', 'sata5', 'scsi0', 'scsi1', 'scsi2', 'scsi3', 'scsi4', 'scsi5', 'scsi6', 'scsi7', 'scsi8', 'scsi9', 'scsi10', 'scsi11', 'scsi12', 'scsi13', 'scsi14', 'scsi15', 'scsi16', 'scsi17', 'scsi18', 'scsi19', 'scsi20', 'scsi21', 'scsi22', 'scsi23', 'scsi24', 'scsi25', 'scsi26', 'scsi27', 'scsi28', 'scsi29', 'scsi30', 'virtio0', 'virtio1', 'virtio2', 'virtio3', 'virtio4', 'virtio5', 'virtio6', 'virtio7', 'virtio8', 'virtio9', 'virtio10', 'virtio11', 'virtio12', 'virtio13', 'virtio14', 'virtio15'")
					}},
				"storage": subSchemaDiskStorage(schema.Schema{Optional: true}),
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "disk",
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Errorf(errorMSG.String, k)
						}
						switch v {
						case "disk", "cdrom", "cloudinit":
							return nil
						}
						return diag.Errorf("type must be one of 'disk', 'cdrom', 'cloudinit'")
					}},
				"wwn": subSchemaDiskWWN(),
			}}}
}

func SchemaDisks() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ide": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ide0": subSchemaIde("ide0"),
							"ide1": subSchemaIde("ide1"),
							"ide2": subSchemaIde("ide2"),
							"ide3": subSchemaIde("ide3")}}},
				"sata": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sata0": subSchemaSata("sata0"),
							"sata1": subSchemaSata("sata1"),
							"sata2": subSchemaSata("sata2"),
							"sata3": subSchemaSata("sata3"),
							"sata4": subSchemaSata("sata4"),
							"sata5": subSchemaSata("sata5")}}},
				"scsi": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"scsi0":  subSchemaScsi("scsi0"),
							"scsi1":  subSchemaScsi("scsi1"),
							"scsi2":  subSchemaScsi("scsi2"),
							"scsi3":  subSchemaScsi("scsi3"),
							"scsi4":  subSchemaScsi("scsi4"),
							"scsi5":  subSchemaScsi("scsi5"),
							"scsi6":  subSchemaScsi("scsi6"),
							"scsi7":  subSchemaScsi("scsi7"),
							"scsi8":  subSchemaScsi("scsi8"),
							"scsi9":  subSchemaScsi("scsi9"),
							"scsi10": subSchemaScsi("scsi10"),
							"scsi11": subSchemaScsi("scsi11"),
							"scsi12": subSchemaScsi("scsi12"),
							"scsi13": subSchemaScsi("scsi13"),
							"scsi14": subSchemaScsi("scsi14"),
							"scsi15": subSchemaScsi("scsi15"),
							"scsi16": subSchemaScsi("scsi16"),
							"scsi17": subSchemaScsi("scsi17"),
							"scsi18": subSchemaScsi("scsi18"),
							"scsi19": subSchemaScsi("scsi19"),
							"scsi20": subSchemaScsi("scsi20"),
							"scsi21": subSchemaScsi("scsi21"),
							"scsi22": subSchemaScsi("scsi22"),
							"scsi23": subSchemaScsi("scsi23"),
							"scsi24": subSchemaScsi("scsi24"),
							"scsi25": subSchemaScsi("scsi25"),
							"scsi26": subSchemaScsi("scsi26"),
							"scsi27": subSchemaScsi("scsi27"),
							"scsi28": subSchemaScsi("scsi28"),
							"scsi29": subSchemaScsi("scsi29"),
							"scsi30": subSchemaScsi("scsi30")}}},
				"virtio": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"virtio0":  subSchemaVirtio("virtio0"),
							"virtio1":  subSchemaVirtio("virtio1"),
							"virtio2":  subSchemaVirtio("virtio2"),
							"virtio3":  subSchemaVirtio("virtio3"),
							"virtio4":  subSchemaVirtio("virtio4"),
							"virtio5":  subSchemaVirtio("virtio5"),
							"virtio6":  subSchemaVirtio("virtio6"),
							"virtio7":  subSchemaVirtio("virtio7"),
							"virtio8":  subSchemaVirtio("virtio8"),
							"virtio9":  subSchemaVirtio("virtio9"),
							"virtio10": subSchemaVirtio("virtio10"),
							"virtio11": subSchemaVirtio("virtio11"),
							"virtio12": subSchemaVirtio("virtio12"),
							"virtio13": subSchemaVirtio("virtio13"),
							"virtio14": subSchemaVirtio("virtio14"),
							"virtio15": subSchemaVirtio("virtio15")}}}}}}
}

func subSchemaCdRom(path string, ci bool) *schema.Schema {
	var conflicts []string
	if ci {
		conflicts = []string{path + ".cloudinit", path + ".disk", path + ".passthrough"}
	} else {
		conflicts = []string{path + ".disk", path + ".passthrough"}
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: conflicts,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"iso": subSchemaIsoPath(schema.Schema{ConflictsWith: []string{path + ".cdrom.0.passthrough"}}),
				"passthrough": {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{path + ".cdrom.0.iso"}}}}}
}

func subSchemaCloudInit(path, slot string) *schema.Schema {
	// 41 is all the disk slots for cloudinit
	// 3 are the conflicts within the same disk slot
	c := append(make([]string, 0, 44), path+".cdrom", path+".disk", path+".passthrough")
	if slot != "ide0" {
		c = append(c, "disks.0.ide.0.ide0.0.cloudinit")
	}
	if slot != "ide1" {
		c = append(c, "disks.0.ide.0.ide1.0.cloudinit")
	}
	if slot != "ide2" {
		c = append(c, "disks.0.ide.0.ide2.0.cloudinit")
	}
	if slot != "ide3" {
		c = append(c, "disks.0.ide.0.ide3.0.cloudinit")
	}
	if slot != "sata0" {
		c = append(c, "disks.0.sata.0.sata0.0.cloudinit")
	}
	if slot != "sata1" {
		c = append(c, "disks.0.sata.0.sata1.0.cloudinit")
	}
	if slot != "sata2" {
		c = append(c, "disks.0.sata.0.sata2.0.cloudinit")
	}
	if slot != "sata3" {
		c = append(c, "disks.0.sata.0.sata3.0.cloudinit")
	}
	if slot != "sata4" {
		c = append(c, "disks.0.sata.0.sata4.0.cloudinit")
	}
	if slot != "sata5" {
		c = append(c, "disks.0.sata.0.sata5.0.cloudinit")
	}
	if slot != "scsi0" {
		c = append(c, "disks.0.scsi.0.scsi0.0.cloudinit")
	}
	if slot != "scsi1" {
		c = append(c, "disks.0.scsi.0.scsi1.0.cloudinit")
	}
	if slot != "scsi2" {
		c = append(c, "disks.0.scsi.0.scsi2.0.cloudinit")
	}
	if slot != "scsi3" {
		c = append(c, "disks.0.scsi.0.scsi3.0.cloudinit")
	}
	if slot != "scsi4" {
		c = append(c, "disks.0.scsi.0.scsi4.0.cloudinit")
	}
	if slot != "scsi5" {
		c = append(c, "disks.0.scsi.0.scsi5.0.cloudinit")
	}
	if slot != "scsi6" {
		c = append(c, "disks.0.scsi.0.scsi6.0.cloudinit")
	}
	if slot != "scsi7" {
		c = append(c, "disks.0.scsi.0.scsi7.0.cloudinit")
	}
	if slot != "scsi8" {
		c = append(c, "disks.0.scsi.0.scsi8.0.cloudinit")
	}
	if slot != "scsi9" {
		c = append(c, "disks.0.scsi.0.scsi9.0.cloudinit")
	}
	if slot != "scsi10" {
		c = append(c, "disks.0.scsi.0.scsi10.0.cloudinit")
	}
	if slot != "scsi11" {
		c = append(c, "disks.0.scsi.0.scsi11.0.cloudinit")
	}
	if slot != "scsi12" {
		c = append(c, "disks.0.scsi.0.scsi12.0.cloudinit")
	}
	if slot != "scsi13" {
		c = append(c, "disks.0.scsi.0.scsi13.0.cloudinit")
	}
	if slot != "scsi14" {
		c = append(c, "disks.0.scsi.0.scsi14.0.cloudinit")
	}
	if slot != "scsi15" {
		c = append(c, "disks.0.scsi.0.scsi15.0.cloudinit")
	}
	if slot != "scsi16" {
		c = append(c, "disks.0.scsi.0.scsi16.0.cloudinit")
	}
	if slot != "scsi17" {
		c = append(c, "disks.0.scsi.0.scsi17.0.cloudinit")
	}
	if slot != "scsi18" {
		c = append(c, "disks.0.scsi.0.scsi18.0.cloudinit")
	}
	if slot != "scsi19" {
		c = append(c, "disks.0.scsi.0.scsi19.0.cloudinit")
	}
	if slot != "scsi20" {
		c = append(c, "disks.0.scsi.0.scsi20.0.cloudinit")
	}
	if slot != "scsi21" {
		c = append(c, "disks.0.scsi.0.scsi21.0.cloudinit")
	}
	if slot != "scsi22" {
		c = append(c, "disks.0.scsi.0.scsi22.0.cloudinit")
	}
	if slot != "scsi23" {
		c = append(c, "disks.0.scsi.0.scsi23.0.cloudinit")
	}
	if slot != "scsi24" {
		c = append(c, "disks.0.scsi.0.scsi24.0.cloudinit")
	}
	if slot != "scsi25" {
		c = append(c, "disks.0.scsi.0.scsi25.0.cloudinit")
	}
	if slot != "scsi26" {
		c = append(c, "disks.0.scsi.0.scsi26.0.cloudinit")
	}
	if slot != "scsi27" {
		c = append(c, "disks.0.scsi.0.scsi27.0.cloudinit")
	}
	if slot != "scsi28" {
		c = append(c, "disks.0.scsi.0.scsi28.0.cloudinit")
	}
	if slot != "scsi29" {
		c = append(c, "disks.0.scsi.0.scsi29.0.cloudinit")
	}
	if slot != "scsi30" {
		c = append(c, "disks.0.scsi.0.scsi30.0.cloudinit")
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: c,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"storage": subSchemaDiskStorage(schema.Schema{Required: true})}}}
}

func subSchemaDiskAsyncIO() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorMSG.String, k)
			}
			if err := pveAPI.QemuDiskAsyncIO(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskBackup() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	}
}

func subSchemaDiskBandwidthIopsBurst() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok || v < 0 {
				return diag.Errorf(errorMSG.Uint, k)
			}
			if err := pveAPI.QemuDiskBandwidthIopsLimitBurst(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskBandwidthIopsBurstLength() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0}
}

func subSchemaDiskBandwidthIopsConcurrent() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok || v < 0 {
				return diag.Errorf(errorMSG.Uint, k)
			}
			if err := pveAPI.QemuDiskBandwidthIopsLimitConcurrent(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskBandwidthMBpsBurst() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(float64)
			if !ok {
				return diag.Errorf(errorMSG.Float, k)
			}
			if err := pveAPI.QemuDiskBandwidthMBpsLimitBurst(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskBandwidthMBpsConcurrent() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(float64)
			if !ok {
				return diag.Errorf(errorMSG.Float, k)
			}
			if err := pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskBandwidth(params map[string]*schema.Schema) map[string]*schema.Schema {
	params["mbps_r_burst"] = subSchemaDiskBandwidthMBpsBurst()
	params["mbps_r_concurrent"] = subSchemaDiskBandwidthMBpsConcurrent()
	params["mbps_wr_burst"] = subSchemaDiskBandwidthMBpsBurst()
	params["mbps_wr_concurrent"] = subSchemaDiskBandwidthMBpsConcurrent()
	params["iops_r_burst"] = subSchemaDiskBandwidthIopsBurst()
	params["iops_r_burst_length"] = subSchemaDiskBandwidthIopsBurstLength()
	params["iops_r_concurrent"] = subSchemaDiskBandwidthIopsConcurrent()
	params["iops_wr_burst"] = subSchemaDiskBandwidthIopsBurst()
	params["iops_wr_burst_length"] = subSchemaDiskBandwidthIopsBurstLength()
	params["iops_wr_concurrent"] = subSchemaDiskBandwidthIopsConcurrent()
	return params
}

func subSchemaDiskCache() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "",
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorMSG.String, k)
			}
			if err := pveAPI.QemuDiskCache(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskFormat(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(errorMSG.String, k)
		}
		if err := pveAPI.QemuDiskFormat(v).Validate(); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	return &s
}

func subSchemaDiskId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true}
}

func subSchemaDiskSerial() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "",
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorMSG.String, k)
			}
			if err := pveAPI.QemuDiskSerial(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaDiskSize(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(errorMSG.String, k)
		}
		if !regexp.MustCompile(`^[123456789]\d*[KMGT]?$`).MatchString(v) {
			return diag.Errorf("%s must match the following regex ^[123456789]\\d*[KMGT]?$", k)
		}
		return nil
	}
	s.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
		return convert_SizeStringToKibibytes_Unsafe(old) == convert_SizeStringToKibibytes_Unsafe(new)
	}
	return &s
}

func subSchemaDiskStorage(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	return &s
}

func subSchemaDiskWWN() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorMSG.String, k)
			}
			if err := pveAPI.QemuWorldWideName(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaIde(slot string) *schema.Schema {
	path := "disks.0.ide.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     subSchemaCdRom(path, true),
				"cloudinit": subSchemaCloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":        subSchemaDiskAsyncIO(),
							"backup":         subSchemaDiskBackup(),
							"cache":          subSchemaDiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							"id":             subSchemaDiskId(),
							"linked_disk_id": subSchemaLinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         subSchemaDiskSerial(),
							"size":           subSchemaDiskSize(schema.Schema{Required: true}),
							"storage":        subSchemaDiskStorage(schema.Schema{Required: true}),
							"wwn":            subSchemaDiskWWN()})}},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":    subSchemaDiskAsyncIO(),
							"backup":     subSchemaDiskBackup(),
							"cache":      subSchemaDiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       subSchemaPassthroughFile(schema.Schema{Required: true}),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     subSchemaDiskSerial(),
							"size":       subSchemaPassthroughSize(),
							"wwn":        subSchemaDiskWWN()})}}}}}
}

func subSchemaIsoPath(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	return &s
}

func subSchemaLinkedDiskId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true}
}

func subSchemaPassthroughFile(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	return &s
}

func subSchemaPassthroughSize() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true}
}

func subSchemaSata(slot string) *schema.Schema {
	path := "disks.0.sata.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     subSchemaCdRom(path, true),
				"cloudinit": subSchemaCloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":        subSchemaDiskAsyncIO(),
							"backup":         subSchemaDiskBackup(),
							"cache":          subSchemaDiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							"id":             subSchemaDiskId(),
							"linked_disk_id": subSchemaLinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         subSchemaDiskSerial(),
							"size":           subSchemaDiskSize(schema.Schema{Required: true}),
							"storage":        subSchemaDiskStorage(schema.Schema{Required: true}),
							"wwn":            subSchemaDiskWWN()})}},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":    subSchemaDiskAsyncIO(),
							"backup":     subSchemaDiskBackup(),
							"cache":      subSchemaDiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       subSchemaPassthroughFile(schema.Schema{Required: true}),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     subSchemaDiskSerial(),
							"size":       subSchemaPassthroughSize(),
							"wwn":        subSchemaDiskWWN()})}}}}}
}

func subSchemaScsi(slot string) *schema.Schema {
	path := "disks.0.scsi.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     subSchemaCdRom(path, true),
				"cloudinit": subSchemaCloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":        subSchemaDiskAsyncIO(),
							"backup":         subSchemaDiskBackup(),
							"cache":          subSchemaDiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							"id":             subSchemaDiskId(),
							"iothread":       {Type: schema.TypeBool, Optional: true},
							"linked_disk_id": subSchemaLinkedDiskId(),
							"readonly":       {Type: schema.TypeBool, Optional: true},
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         subSchemaDiskSerial(),
							"size":           subSchemaDiskSize(schema.Schema{Required: true}),
							"storage":        subSchemaDiskStorage(schema.Schema{Required: true}),
							"wwn":            subSchemaDiskWWN()})}},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":    subSchemaDiskAsyncIO(),
							"backup":     subSchemaDiskBackup(),
							"cache":      subSchemaDiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       subSchemaPassthroughFile(schema.Schema{Required: true}),
							"iothread":   {Type: schema.TypeBool, Optional: true},
							"readonly":   {Type: schema.TypeBool, Optional: true},
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     subSchemaDiskSerial(),
							"size":       subSchemaPassthroughSize(),
							"wwn":        subSchemaDiskWWN()})}}}}}
}

func subSchemaVirtio(setting string) *schema.Schema {
	path := "disks.0.virtio.0." + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom": subSchemaCdRom(path, false),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							"asyncio":        subSchemaDiskAsyncIO(),
							"backup":         subSchemaDiskBackup(),
							"cache":          subSchemaDiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"format":         subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							"id":             subSchemaDiskId(),
							"iothread":       {Type: schema.TypeBool, Optional: true},
							"linked_disk_id": subSchemaLinkedDiskId(),
							"readonly":       {Type: schema.TypeBool, Optional: true},
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         subSchemaDiskSerial(),
							"size":           subSchemaDiskSize(schema.Schema{Required: true}),
							"storage":        subSchemaDiskStorage(schema.Schema{Required: true}),
							"wwn":            subSchemaDiskWWN()})}},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".disk"},
					Elem: &schema.Resource{Schema: subSchemaDiskBandwidth(
						map[string]*schema.Schema{
							"asyncio":   subSchemaDiskAsyncIO(),
							"backup":    subSchemaDiskBackup(),
							"cache":     subSchemaDiskCache(),
							"discard":   {Type: schema.TypeBool, Optional: true},
							"file":      subSchemaPassthroughFile(schema.Schema{Required: true}),
							"iothread":  {Type: schema.TypeBool, Optional: true},
							"readonly":  {Type: schema.TypeBool, Optional: true},
							"replicate": {Type: schema.TypeBool, Optional: true},
							"serial":    subSchemaDiskSerial(),
							"size":      subSchemaPassthroughSize(),
							"wwn":       subSchemaDiskWWN()})}}}}}
}
