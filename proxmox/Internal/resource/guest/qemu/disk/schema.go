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

	schemaAsyncIO           string = "asyncio"
	schemaBackup            string = "backup"
	schemaCache             string = "cache"
	schemaCdRom             string = "cdrom"
	schemaCdRomISO          string = "iso"
	schemaCdRomPassthrough  string = "passthrough"
	schemaCloudInit         string = "cloudinit"
	schemaDiscard           string = "discard"
	schemaDisk              string = "disk"
	schemaDiskFile          string = "disk_file"
	schemaEmulateSSD        string = "emulatessd"
	schemaFile              string = "file"
	schemaFormat            string = "format"
	schemaID                string = "id"
	schemaIDE               string = "ide"
	schemaIOPSrBurst        string = "iops_r_burst"
	schemaIOPSrBurstLength  string = "iops_r_burst_length"
	schemaIOPSrConcurrent   string = "iops_r_concurrent"
	schemaIOPSwrBurst       string = "iops_wr_burst"
	schemaIOPSwrBurstLength string = "iops_wr_burst_length"
	schemaIOPSwrConcurrent  string = "iops_wr_concurrent"
	schemaIOthread          string = "iothread"
	schemaISO               string = "iso"
	schemaLinkedDiskId      string = "linked_disk_id"
	schemaMBPSrBurst        string = "mbps_r_burst"
	schemaMBPSrConcurrent   string = "mbps_r_concurrent"
	schemaMBPSwrBurst       string = "mbps_wr_burst"
	schemaMBPSwrConcurrent  string = "mbps_wr_concurrent"
	schemaPassthrough       string = "passthrough"
	schemaReadOnly          string = "readonly"
	schemaReplicate         string = "replicate"
	schemaSata              string = "sata"
	schemaScsi              string = "scsi"
	schemaSerial            string = "serial"
	schemaSize              string = "size"
	schemaSlot              string = "slot"
	schemaStorage           string = "storage"
	schemaType              string = "type"
	schemaVirtIO            string = "virtio"
	schemaWorldWideName     string = "wwn"

	pathIDE    string = RootDisks + ".0." + schemaIDE + ".0."
	pathSata   string = RootDisks + ".0." + schemaSata + ".0."
	pathScsi   string = RootDisks + ".0." + schemaScsi + ".0."
	pathVirtIO string = RootDisks + ".0." + schemaVirtIO + ".0."

	enumCdRom     string = "cdrom"
	enumCloudInit string = "cloudinit"
	enumDisk      string = "disk"

	slotIDE    string = schemaIDE
	slotSata   string = schemaSata
	slotScsi   string = schemaScsi
	slotVirtIO string = schemaVirtIO
)

func SchemaDisk() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{RootDisks},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaAsyncIO:           subSchemaDiskAsyncIO(),
				schemaBackup:            subSchemaDiskBackup(),
				schemaCache:             subSchemaDiskCache(),
				schemaDiscard:           {Type: schema.TypeBool, Optional: true},
				schemaDiskFile:          subSchemaPassthroughFile(schema.Schema{Optional: true}),
				schemaEmulateSSD:        {Type: schema.TypeBool, Optional: true},
				schemaFormat:            subSchemaDiskFormat(schema.Schema{}),
				schemaID:                subSchemaDiskId(),
				schemaIOPSrBurst:        subSchemaDiskBandwidthIopsBurst(),
				schemaIOPSrBurstLength:  subSchemaDiskBandwidthIopsBurstLength(),
				schemaIOPSrConcurrent:   subSchemaDiskBandwidthIopsConcurrent(),
				schemaIOPSwrBurst:       subSchemaDiskBandwidthIopsBurst(),
				schemaIOPSwrBurstLength: subSchemaDiskBandwidthIopsBurstLength(),
				schemaIOPSwrConcurrent:  subSchemaDiskBandwidthIopsConcurrent(),
				schemaIOthread:          {Type: schema.TypeBool, Optional: true},
				schemaISO:               subSchemaIsoPath(schema.Schema{}),
				schemaLinkedDiskId:      subSchemaLinkedDiskId(),
				schemaMBPSrBurst:        subSchemaDiskBandwidthMBpsBurst(),
				schemaMBPSrConcurrent:   subSchemaDiskBandwidthMBpsConcurrent(),
				schemaMBPSwrBurst:       subSchemaDiskBandwidthMBpsBurst(),
				schemaMBPSwrConcurrent:  subSchemaDiskBandwidthMBpsConcurrent(),
				schemaPassthrough:       {Type: schema.TypeBool, Default: false, Optional: true},
				schemaReadOnly:          {Type: schema.TypeBool, Optional: true},
				schemaReplicate:         {Type: schema.TypeBool, Optional: true},
				schemaSerial:            subSchemaDiskSerial(),
				schemaSize:              subSchemaDiskSize(schema.Schema{Computed: true, Optional: true}),
				schemaSlot: {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Errorf(errorMSG.String, k)
						}
						switch v {
						case slotIDE + "0", slotIDE + "1", slotIDE + "2",
							slotSata + "0", slotSata + "1", slotSata + "2", slotSata + "3", slotSata + "4", slotSata + "5",
							slotScsi + "0", slotScsi + "1", slotScsi + "2", slotScsi + "3", slotScsi + "4", slotScsi + "5", slotScsi + "6", slotScsi + "7", slotScsi + "8", slotScsi + "9", slotScsi + "10", slotScsi + "11", slotScsi + "12", slotScsi + "13", slotScsi + "14", slotScsi + "15", slotScsi + "16", slotScsi + "17", slotScsi + "18", slotScsi + "19", slotScsi + "20", slotScsi + "21", slotScsi + "22", slotScsi + "23", slotScsi + "24", slotScsi + "25", slotScsi + "26", slotScsi + "27", slotScsi + "28", slotScsi + "29", slotScsi + "30",
							slotVirtIO + "0", slotVirtIO + "1", slotVirtIO + "2", slotVirtIO + "3", slotVirtIO + "4", slotVirtIO + "5", slotVirtIO + "6", slotVirtIO + "7", slotVirtIO + "8", slotVirtIO + "9", slotVirtIO + "10", slotVirtIO + "11", slotVirtIO + "12", slotVirtIO + "13", slotVirtIO + "14", slotVirtIO + "15":
							return nil
						}
						return diag.Errorf(schemaSlot + " must be one of '" + slotIDE + "0', '" + slotIDE + "1', '" + slotIDE + "2', '" + slotSata + "0', '" + slotSata + "1', '" + slotSata + "2', '" + slotSata + "3', '" + slotSata + "4', '" + slotSata + "5', '" + slotScsi + "0', '" + slotScsi + "1', '" + slotScsi + "2', '" + slotScsi + "3', '" + slotScsi + "4', '" + slotScsi + "5', '" + slotScsi + "6', '" + slotScsi + "7', '" + slotScsi + "8', '" + slotScsi + "9', '" + slotScsi + "10', '" + slotScsi + "11', '" + slotScsi + "12', '" + slotScsi + "13', '" + slotScsi + "14', '" + slotScsi + "15', '" + slotScsi + "16', '" + slotScsi + "17', '" + slotScsi + "18', '" + slotScsi + "19', '" + slotScsi + "20', '" + slotScsi + "21', '" + slotScsi + "22', '" + slotScsi + "23', '" + slotScsi + "24', '" + slotScsi + "25', '" + slotScsi + "26', '" + slotScsi + "27', '" + slotScsi + "28', '" + slotScsi + "29', '" + slotScsi + "30', '" + slotVirtIO + "0', '" + slotVirtIO + "1', '" + slotVirtIO + "2', '" + slotVirtIO + "3', '" + slotVirtIO + "4', '" + slotVirtIO + "5', '" + slotVirtIO + "6', '" + slotVirtIO + "7', '" + slotVirtIO + "8', '" + slotVirtIO + "9', '" + slotVirtIO + "10', '" + slotVirtIO + "11', '" + slotVirtIO + "12', '" + slotVirtIO + "13', '" + slotVirtIO + "14', '" + slotVirtIO + "15'")
					}},
				schemaStorage: subSchemaDiskStorage(schema.Schema{Optional: true}),
				schemaType: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  schemaDisk,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Errorf(errorMSG.String, k)
						}
						switch v {
						case schemaDisk, schemaCdRom, schemaCloudInit:
							return nil
						}
						return diag.Errorf(schemaType + " must be one of '" + enumDisk + "', '" + enumCdRom + "', '" + enumCloudInit + "'")
					}},
				schemaWorldWideName: subSchemaDiskWWN(),
			}}}
}

func SchemaDisks() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaIDE: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaIDE + "0": subSchemaIde(schemaIDE + "0"),
							schemaIDE + "1": subSchemaIde(schemaIDE + "1"),
							schemaIDE + "2": subSchemaIde(schemaIDE + "2"),
							schemaIDE + "3": subSchemaIde(schemaIDE + "3")}}},
				schemaSata: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaSata + "0": subSchemaSata(schemaSata + "0"),
							schemaSata + "1": subSchemaSata(schemaSata + "1"),
							schemaSata + "2": subSchemaSata(schemaSata + "2"),
							schemaSata + "3": subSchemaSata(schemaSata + "3"),
							schemaSata + "4": subSchemaSata(schemaSata + "4"),
							schemaSata + "5": subSchemaSata(schemaSata + "5")}}},
				schemaScsi: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaScsi + "0":  subSchemaScsi(schemaScsi + "0"),
							schemaScsi + "1":  subSchemaScsi(schemaScsi + "1"),
							schemaScsi + "2":  subSchemaScsi(schemaScsi + "2"),
							schemaScsi + "3":  subSchemaScsi(schemaScsi + "3"),
							schemaScsi + "4":  subSchemaScsi(schemaScsi + "4"),
							schemaScsi + "5":  subSchemaScsi(schemaScsi + "5"),
							schemaScsi + "6":  subSchemaScsi(schemaScsi + "6"),
							schemaScsi + "7":  subSchemaScsi(schemaScsi + "7"),
							schemaScsi + "8":  subSchemaScsi(schemaScsi + "8"),
							schemaScsi + "9":  subSchemaScsi(schemaScsi + "9"),
							schemaScsi + "10": subSchemaScsi(schemaScsi + "10"),
							schemaScsi + "11": subSchemaScsi(schemaScsi + "11"),
							schemaScsi + "12": subSchemaScsi(schemaScsi + "12"),
							schemaScsi + "13": subSchemaScsi(schemaScsi + "13"),
							schemaScsi + "14": subSchemaScsi(schemaScsi + "14"),
							schemaScsi + "15": subSchemaScsi(schemaScsi + "15"),
							schemaScsi + "16": subSchemaScsi(schemaScsi + "16"),
							schemaScsi + "17": subSchemaScsi(schemaScsi + "17"),
							schemaScsi + "18": subSchemaScsi(schemaScsi + "18"),
							schemaScsi + "19": subSchemaScsi(schemaScsi + "19"),
							schemaScsi + "20": subSchemaScsi(schemaScsi + "20"),
							schemaScsi + "21": subSchemaScsi(schemaScsi + "21"),
							schemaScsi + "22": subSchemaScsi(schemaScsi + "22"),
							schemaScsi + "23": subSchemaScsi(schemaScsi + "23"),
							schemaScsi + "24": subSchemaScsi(schemaScsi + "24"),
							schemaScsi + "25": subSchemaScsi(schemaScsi + "25"),
							schemaScsi + "26": subSchemaScsi(schemaScsi + "26"),
							schemaScsi + "27": subSchemaScsi(schemaScsi + "27"),
							schemaScsi + "28": subSchemaScsi(schemaScsi + "28"),
							schemaScsi + "29": subSchemaScsi(schemaScsi + "29"),
							schemaScsi + "30": subSchemaScsi(schemaScsi + "30")}}},
				schemaVirtIO: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaVirtIO + "0":  subSchemaVirtio(schemaVirtIO + "0"),
							schemaVirtIO + "1":  subSchemaVirtio(schemaVirtIO + "1"),
							schemaVirtIO + "2":  subSchemaVirtio(schemaVirtIO + "2"),
							schemaVirtIO + "3":  subSchemaVirtio(schemaVirtIO + "3"),
							schemaVirtIO + "4":  subSchemaVirtio(schemaVirtIO + "4"),
							schemaVirtIO + "5":  subSchemaVirtio(schemaVirtIO + "5"),
							schemaVirtIO + "6":  subSchemaVirtio(schemaVirtIO + "6"),
							schemaVirtIO + "7":  subSchemaVirtio(schemaVirtIO + "7"),
							schemaVirtIO + "8":  subSchemaVirtio(schemaVirtIO + "8"),
							schemaVirtIO + "9":  subSchemaVirtio(schemaVirtIO + "9"),
							schemaVirtIO + "10": subSchemaVirtio(schemaVirtIO + "10"),
							schemaVirtIO + "11": subSchemaVirtio(schemaVirtIO + "11"),
							schemaVirtIO + "12": subSchemaVirtio(schemaVirtIO + "12"),
							schemaVirtIO + "13": subSchemaVirtio(schemaVirtIO + "13"),
							schemaVirtIO + "14": subSchemaVirtio(schemaVirtIO + "14"),
							schemaVirtIO + "15": subSchemaVirtio(schemaVirtIO + "15")}}}}}}
}

func subSchemaCdRom(path string, ci bool) *schema.Schema {
	var conflicts []string
	if ci {
		conflicts = []string{path + "." + schemaCloudInit, path + "." + schemaDisk, path + "." + schemaPassthrough}
	} else {
		conflicts = []string{path + "." + schemaDisk, path + "." + schemaPassthrough}
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: conflicts,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCdRomISO: subSchemaIsoPath(schema.Schema{ConflictsWith: []string{path + "." + schemaCdRom + ".0." + schemaCdRomPassthrough}}),
				schemaCdRomPassthrough: {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{path + "." + schemaCdRom + ".0." + schemaCdRomISO}}}}}
}

func subSchemaCloudInit(path, slot string) *schema.Schema {
	// 41 is all the disk slots for cloudinit
	// 3 are the conflicts within the same disk slot
	c := append(make([]string, 0, 44), path+"."+schemaCdRom, path+"."+schemaDisk, path+"."+schemaPassthrough)
	if slot != schemaIDE+"0" {
		c = append(c, pathIDE+schemaIDE+"0.0."+schemaCloudInit)
	}
	if slot != schemaIDE+"1" {
		c = append(c, pathIDE+schemaIDE+"1.0."+schemaCloudInit)
	}
	if slot != schemaIDE+"2" {
		c = append(c, pathIDE+schemaIDE+"2.0."+schemaCloudInit)
	}
	if slot != schemaIDE+"3" {
		c = append(c, pathIDE+schemaIDE+"3.0."+schemaCloudInit)
	}
	if slot != schemaSata+"0" {
		c = append(c, pathSata+schemaSata+"0.0."+schemaCloudInit)
	}
	if slot != schemaSata+"1" {
		c = append(c, pathSata+schemaSata+"1.0."+schemaCloudInit)
	}
	if slot != schemaSata+"2" {
		c = append(c, pathSata+schemaSata+"2.0."+schemaCloudInit)
	}
	if slot != schemaSata+"3" {
		c = append(c, pathSata+schemaSata+"3.0."+schemaCloudInit)
	}
	if slot != schemaSata+"4" {
		c = append(c, pathSata+schemaSata+"4.0."+schemaCloudInit)
	}
	if slot != schemaSata+"5" {
		c = append(c, pathSata+schemaSata+"5.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"0" {
		c = append(c, pathScsi+schemaScsi+"0.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"1" {
		c = append(c, pathScsi+schemaScsi+"1.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"2" {
		c = append(c, pathScsi+schemaScsi+"2.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"3" {
		c = append(c, pathScsi+schemaScsi+"3.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"4" {
		c = append(c, pathScsi+schemaScsi+"4.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"5" {
		c = append(c, pathScsi+schemaScsi+"5.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"6" {
		c = append(c, pathScsi+schemaScsi+"6.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"7" {
		c = append(c, pathScsi+schemaScsi+"7.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"8" {
		c = append(c, pathScsi+schemaScsi+"8.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"9" {
		c = append(c, pathScsi+schemaScsi+"9.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"10" {
		c = append(c, pathScsi+schemaScsi+"10.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"11" {
		c = append(c, pathScsi+schemaScsi+"11.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"12" {
		c = append(c, pathScsi+schemaScsi+"12.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"13" {
		c = append(c, pathScsi+schemaScsi+"13.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"14" {
		c = append(c, pathScsi+schemaScsi+"14.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"15" {
		c = append(c, pathScsi+schemaScsi+"15.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"16" {
		c = append(c, pathScsi+schemaScsi+"16.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"17" {
		c = append(c, pathScsi+schemaScsi+"17.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"18" {
		c = append(c, pathScsi+schemaScsi+"18.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"19" {
		c = append(c, pathScsi+schemaScsi+"19.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"20" {
		c = append(c, pathScsi+schemaScsi+"20.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"21" {
		c = append(c, pathScsi+schemaScsi+"21.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"22" {
		c = append(c, pathScsi+schemaScsi+"22.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"23" {
		c = append(c, pathScsi+schemaScsi+"23.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"24" {
		c = append(c, pathScsi+schemaScsi+"24.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"25" {
		c = append(c, pathScsi+schemaScsi+"25.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"26" {
		c = append(c, pathScsi+schemaScsi+"26.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"27" {
		c = append(c, pathScsi+schemaScsi+"27.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"28" {
		c = append(c, pathScsi+schemaScsi+"28.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"29" {
		c = append(c, pathScsi+schemaScsi+"29.0."+schemaCloudInit)
	}
	if slot != schemaScsi+"30" {
		c = append(c, pathScsi+schemaScsi+"30.0."+schemaCloudInit)
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: c,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaStorage: subSchemaDiskStorage(schema.Schema{Required: true})}}}
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
	params[schemaMBPSrBurst] = subSchemaDiskBandwidthMBpsBurst()
	params[schemaMBPSrConcurrent] = subSchemaDiskBandwidthMBpsConcurrent()
	params[schemaMBPSwrBurst] = subSchemaDiskBandwidthMBpsBurst()
	params[schemaMBPSwrConcurrent] = subSchemaDiskBandwidthMBpsConcurrent()
	params[schemaIOPSrBurst] = subSchemaDiskBandwidthIopsBurst()
	params[schemaIOPSrBurstLength] = subSchemaDiskBandwidthIopsBurstLength()
	params[schemaIOPSrConcurrent] = subSchemaDiskBandwidthIopsConcurrent()
	params[schemaIOPSwrBurst] = subSchemaDiskBandwidthIopsBurst()
	params[schemaIOPSwrBurstLength] = subSchemaDiskBandwidthIopsBurstLength()
	params[schemaIOPSwrConcurrent] = subSchemaDiskBandwidthIopsConcurrent()
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
	path := pathIDE + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCdRom:     subSchemaCdRom(path, true),
				schemaCloudInit: subSchemaCloudInit(path, slot),
				schemaDisk: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaPassthrough},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFormat:        subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							schemaID:            subSchemaDiskId(),
							schemaLinkedDiskId:  subSchemaLinkedDiskId(),
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaDiskSize(schema.Schema{Required: true}),
							schemaStorage:       subSchemaDiskStorage(schema.Schema{Required: true}),
							schemaWorldWideName: subSchemaDiskWWN()})}},
				schemaPassthrough: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaDisk},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFile:          subSchemaPassthroughFile(schema.Schema{Required: true}),
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaPassthroughSize(),
							schemaWorldWideName: subSchemaDiskWWN()})}}}}}
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
	path := pathSata + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCdRom:     subSchemaCdRom(path, true),
				schemaCloudInit: subSchemaCloudInit(path, slot),
				schemaDisk: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaPassthrough},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFormat:        subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							schemaID:            subSchemaDiskId(),
							schemaLinkedDiskId:  subSchemaLinkedDiskId(),
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaDiskSize(schema.Schema{Required: true}),
							schemaStorage:       subSchemaDiskStorage(schema.Schema{Required: true}),
							schemaWorldWideName: subSchemaDiskWWN()})}},
				schemaPassthrough: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaDisk},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFile:          subSchemaPassthroughFile(schema.Schema{Required: true}),
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaPassthroughSize(),
							schemaWorldWideName: subSchemaDiskWWN()})}}}}}
}

func subSchemaScsi(slot string) *schema.Schema {
	path := pathScsi + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCdRom:     subSchemaCdRom(path, true),
				schemaCloudInit: subSchemaCloudInit(path, slot),
				schemaDisk: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaPassthrough},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFormat:        subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							schemaID:            subSchemaDiskId(),
							schemaIOthread:      {Type: schema.TypeBool, Optional: true},
							schemaLinkedDiskId:  subSchemaLinkedDiskId(),
							schemaReadOnly:      {Type: schema.TypeBool, Optional: true},
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaDiskSize(schema.Schema{Required: true}),
							schemaStorage:       subSchemaDiskStorage(schema.Schema{Required: true}),
							schemaWorldWideName: subSchemaDiskWWN()})}},
				schemaPassthrough: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaCloudInit, path + "." + schemaDisk},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaEmulateSSD:    {Type: schema.TypeBool, Optional: true},
							schemaFile:          subSchemaPassthroughFile(schema.Schema{Required: true}),
							schemaIOthread:      {Type: schema.TypeBool, Optional: true},
							schemaReadOnly:      {Type: schema.TypeBool, Optional: true},
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaPassthroughSize(),
							schemaWorldWideName: subSchemaDiskWWN()})}}}}}
}

func subSchemaVirtio(setting string) *schema.Schema {
	path := pathVirtIO + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCdRom: subSchemaCdRom(path, false),
				schemaDisk: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaPassthrough},
					Elem: &schema.Resource{
						Schema: subSchemaDiskBandwidth(map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaFormat:        subSchemaDiskFormat(schema.Schema{Default: "raw"}),
							schemaID:            subSchemaDiskId(),
							schemaIOthread:      {Type: schema.TypeBool, Optional: true},
							schemaLinkedDiskId:  subSchemaLinkedDiskId(),
							schemaReadOnly:      {Type: schema.TypeBool, Optional: true},
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaDiskSize(schema.Schema{Required: true}),
							schemaStorage:       subSchemaDiskStorage(schema.Schema{Required: true}),
							schemaWorldWideName: subSchemaDiskWWN()})}},
				schemaPassthrough: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + "." + schemaCdRom, path + "." + schemaDisk},
					Elem: &schema.Resource{Schema: subSchemaDiskBandwidth(
						map[string]*schema.Schema{
							schemaAsyncIO:       subSchemaDiskAsyncIO(),
							schemaBackup:        subSchemaDiskBackup(),
							schemaCache:         subSchemaDiskCache(),
							schemaDiscard:       {Type: schema.TypeBool, Optional: true},
							schemaFile:          subSchemaPassthroughFile(schema.Schema{Required: true}),
							schemaIOthread:      {Type: schema.TypeBool, Optional: true},
							schemaReadOnly:      {Type: schema.TypeBool, Optional: true},
							schemaReplicate:     {Type: schema.TypeBool, Optional: true},
							schemaSerial:        subSchemaDiskSerial(),
							schemaSize:          subSchemaPassthroughSize(),
							schemaWorldWideName: subSchemaDiskWWN()})}}}}}
}
