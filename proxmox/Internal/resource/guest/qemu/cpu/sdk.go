package cpu

import (
	"errors"
	"strconv"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	error_AffinityFormat = "invalid affinity format"
)

func SDK(d *schema.ResourceData) *pveSDK.QemuCPU {
	v, ok := d.GetOk(Root)
	if !ok { // defaults
		if v := sdkLegacy(d); v != nil {
			return v
		}
		return defaults()
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return nil
	}
	if settings, ok := vv[0].(map[string]any); ok {
		affinity, _ := sdkAffinity(settings[schemaAffinity].(string))
		return &pveSDK.QemuCPU{
			Affinity:     affinity,
			Cores:        util.Pointer(pveSDK.QemuCpuCores(settings[schemaCores].(int))),
			Flags:        sdkFlags(settings),
			Limit:        util.Pointer(pveSDK.CpuLimit(settings[schemaLimit].(int))),
			Numa:         util.Pointer(settings[schemaNuma].(bool)),
			Sockets:      util.Pointer(pveSDK.QemuCpuSockets(settings[schemaSockets].(int))),
			Type:         util.Pointer(pveSDK.CpuType(settings[schemaType].(string))),
			Units:        util.Pointer(pveSDK.CpuUnits(settings[schemaUnits].(int))),
			VirtualCores: util.Pointer(pveSDK.CpuVirtualCores(settings[schemaVirtualCores].(int)))}
	}
	return defaults()
}

func sdkAffinity(rawAffinity string) (*[]uint, error) {
	affinity := make([]uint, 0)
	if rawAffinity == "" {
		return &affinity, nil
	}
	affinityParts := strings.Split(rawAffinity, ",")
	for i := range affinityParts {
		if affinityParts[i] == "" {
			continue
		}
		parts := strings.Split(affinityParts[i], "-")
		switch len(parts) {
		case 1:
			i, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, errors.New(error_AffinityFormat)
			}
			affinity = append(affinity, uint(i))
		case 2:
			start, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, errors.New(error_AffinityFormat)
			}
			end, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, errors.New(error_AffinityFormat)
			}
			if start >= end {
				return nil, errors.New(error_AffinityFormat)
			}
			tmpAffinities := make([]uint, end-start+1)
			for i := start; i <= end; i++ {
				tmpAffinities[i-start] = uint(i)
			}
			affinity = append(affinity, tmpAffinities...)
		default:
			return nil, errors.New(error_AffinityFormat)
		}
	}
	return &affinity, nil
}

func sdkFlags(schema map[string]any) *pveSDK.CpuFlags {
	v, ok := schema[schemaFlags].([]any)
	if !ok || len(v) != 1 || v[0] == nil {
		return defaultFlags()
	}
	schemaItems := v[0].(map[string]any)
	return &pveSDK.CpuFlags{
		AES:        sdkFlag(schemaItems[schemaFlagAes].(string)),
		AmdNoSSB:   sdkFlag(schemaItems[schemaFlagAmdNoSsb].(string)),
		AmdSSBD:    sdkFlag(schemaItems[schemaFlagAmdSsbd].(string)),
		HvEvmcs:    sdkFlag(schemaItems[schemaFlagHvEvmcs].(string)),
		HvTlbFlush: sdkFlag(schemaItems[schemaFlagHvTlbflush].(string)),
		Ibpb:       sdkFlag(schemaItems[schemaFlagIbpb].(string)),
		MdClear:    sdkFlag(schemaItems[schemaFlagMdClear].(string)),
		PCID:       sdkFlag(schemaItems[schemaFlagPcidev].(string)),
		Pdpe1GB:    sdkFlag(schemaItems[schemaFlagPbpe1gb].(string)),
		SSBD:       sdkFlag(schemaItems[schemaFlagSsbd].(string)),
		SpecCtrl:   sdkFlag(schemaItems[schemaFlagSpecCtrl].(string)),
		VirtSSBD:   sdkFlag(schemaItems[schemaFlagVirtSsbd].(string))}
}

func sdkFlag(rawFlag string) *pveSDK.TriBool {
	switch strings.ToLower(rawFlag) {
	case flagOn:
		return util.Pointer(pveSDK.TriBoolTrue)
	case flagOff:
		return util.Pointer(pveSDK.TriBoolFalse)
	}
	return util.Pointer(pveSDK.TriBoolNone)
}

func defaultFlags() *pveSDK.CpuFlags {
	return &pveSDK.CpuFlags{
		AES:        util.Pointer(pveSDK.TriBoolNone),
		AmdNoSSB:   util.Pointer(pveSDK.TriBoolNone),
		AmdSSBD:    util.Pointer(pveSDK.TriBoolNone),
		HvEvmcs:    util.Pointer(pveSDK.TriBoolNone),
		HvTlbFlush: util.Pointer(pveSDK.TriBoolNone),
		Ibpb:       util.Pointer(pveSDK.TriBoolNone),
		MdClear:    util.Pointer(pveSDK.TriBoolNone),
		PCID:       util.Pointer(pveSDK.TriBoolNone),
		Pdpe1GB:    util.Pointer(pveSDK.TriBoolNone),
		SSBD:       util.Pointer(pveSDK.TriBoolNone),
		SpecCtrl:   util.Pointer(pveSDK.TriBoolNone),
		VirtSSBD:   util.Pointer(pveSDK.TriBoolNone)}
}

func defaults() *pveSDK.QemuCPU {
	return &pveSDK.QemuCPU{
		Affinity:     util.Pointer([]uint{}),
		Cores:        util.Pointer(pveSDK.QemuCpuCores(defaultCores)),
		Flags:        defaultFlags(),
		Limit:        util.Pointer(pveSDK.CpuLimit(defaultLimit)),
		Numa:         util.Pointer(defaultNuma),
		Sockets:      util.Pointer(pveSDK.QemuCpuSockets(defaultSockets)),
		Type:         util.Pointer(pveSDK.CpuType(defaultType)),
		Units:        util.Pointer(pveSDK.CpuUnits(defaultUnits)),
		VirtualCores: util.Pointer(pveSDK.CpuVirtualCores(defaultVirtualCores))}
}
