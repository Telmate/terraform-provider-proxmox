package cpu

import (
	"strconv"
	"strings"

	"slices"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveSDK.QemuCPU, d *schema.ResourceData) {
	if _, ok := d.GetOk(Root); !ok {
		if hasLegacySettings := terraformLegacy(config, d); hasLegacySettings {
			return
		}
	}
	cpu := map[string]any{}
	if config.Affinity != nil {
		cpu[schemaAffinity] = terraformAffinity(*config.Affinity)
	} else {
		cpu[schemaAffinity] = defaultAffinity
	}
	if config.Cores != nil {
		cpu[schemaCores] = int(*config.Cores)
	} else {
		cpu[schemaCores] = defaultCores
	}
	if config.Flags != nil {
		if flags := terraformFlags(*config.Flags); flags != nil {
			cpu[schemaFlags] = flags
		}
	}
	if config.Limit != nil {
		cpu[schemaLimit] = int(*config.Limit)
	} else {
		cpu[schemaLimit] = defaultLimit
	}
	if config.Numa != nil {
		cpu[schemaNuma] = *config.Numa
	} else {
		cpu[schemaNuma] = defaultNuma
	}
	if config.Sockets != nil {
		cpu[schemaSockets] = int(*config.Sockets)
	} else {
		cpu[schemaSockets] = defaultSockets
	}
	if config.Type != nil {
		cpu[schemaType] = string(*config.Type)
	} else {
		cpu[schemaType] = defaultType
	}
	if config.Units != nil {
		cpu[schemaUnits] = int(*config.Units)
	} else {
		cpu[schemaUnits] = defaultUnits
	}
	if config.VirtualCores != nil {
		cpu[schemaVirtualCores] = int(*config.VirtualCores)
	} else {
		cpu[schemaVirtualCores] = defaultVirtualCores
	}
	d.Set(Root, []any{cpu})
	terraformLegacyClear(d)
}

func terraformAffinity(affinity []uint) string {
	slices.Sort(affinity)
	var builder strings.Builder
	rangeStart, rangeEnd := affinity[0], affinity[0]
	for i := 1; i < len(affinity); i++ {
		if affinity[i] == affinity[i-1] {
			continue
		}
		if affinity[i] == rangeEnd+1 {
			// Continue the range
			rangeEnd = affinity[i]
		} else {
			// Close the current range and start a new range
			if rangeStart == rangeEnd {
				builder.WriteString(strconv.Itoa(int(rangeStart)) + ",")
			} else {
				builder.WriteString(strconv.Itoa(int(rangeStart)) + "-" + strconv.Itoa(int(rangeEnd)) + ",")
			}
			rangeStart, rangeEnd = affinity[i], affinity[i]
		}
	}
	// Append the last range
	if rangeStart == rangeEnd {
		builder.WriteString(strconv.Itoa(int(rangeStart)))
	} else {
		builder.WriteString(strconv.Itoa(int(rangeStart)) + "-" + strconv.Itoa(int(rangeEnd)))
	}
	return builder.String()
}

func terraformFlag(flag *pveSDK.TriBool) string {
	if flag == nil {
		return ""
	}
	switch *flag {
	case pveSDK.TriBoolTrue:
		return flagOn
	case pveSDK.TriBoolFalse:
		return flagOff
	default:
		return ""
	}
}

func terraformFlags(flags pveSDK.CpuFlags) []map[string]any {
	tmpFlags := [12]string{"", "", "", "", "", "", "", "", "", "", "", ""}
	tmpFlags[0] = terraformFlag(flags.AES)
	tmpFlags[1] = terraformFlag(flags.AmdNoSSB)
	tmpFlags[2] = terraformFlag(flags.AmdSSBD)
	tmpFlags[3] = terraformFlag(flags.HvEvmcs)
	tmpFlags[4] = terraformFlag(flags.HvTlbFlush)
	tmpFlags[5] = terraformFlag(flags.Ibpb)
	tmpFlags[6] = terraformFlag(flags.MdClear)
	tmpFlags[7] = terraformFlag(flags.PCID)
	tmpFlags[8] = terraformFlag(flags.Pdpe1GB)
	tmpFlags[9] = terraformFlag(flags.SSBD)
	tmpFlags[10] = terraformFlag(flags.SpecCtrl)
	tmpFlags[11] = terraformFlag(flags.VirtSSBD)
	for i := range tmpFlags {
		if tmpFlags[i] != "" {
			return []map[string]any{{
				schemaFlagAes:        tmpFlags[0],
				schemaFlagAmdNoSsb:   tmpFlags[1],
				schemaFlagAmdSsbd:    tmpFlags[2],
				schemaFlagHvEvmcs:    tmpFlags[3],
				schemaFlagHvTlbflush: tmpFlags[4],
				schemaFlagIbpb:       tmpFlags[5],
				schemaFlagMdClear:    tmpFlags[6],
				schemaFlagPcidev:     tmpFlags[7],
				schemaFlagPbpe1gb:    tmpFlags[8],
				schemaFlagSsbd:       tmpFlags[9],
				schemaFlagSpecCtrl:   tmpFlags[10],
				schemaFlagVirtSsbd:   tmpFlags[11]}}
		}
	}
	return nil
}
