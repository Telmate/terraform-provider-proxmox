package networks

import pveSDK "github.com/Telmate/proxmox-api-go/proxmox"

func HostManagedVersion() pveSDK.EncodedVersion { return pveSDK.Version{Major: 9, Minor: 1}.Encode() }
