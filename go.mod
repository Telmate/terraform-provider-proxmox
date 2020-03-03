module github.com/Telmate/terraform-provider-proxmox

go 1.13

require (
	github.com/Telmate/proxmox-api-go v0.0.0-20191217000250-7338ae30b9b0
	github.com/hashicorp/terraform v0.12.10
)
replace github.com/Telmate/proxmox-api-go => github.com/claudusd/proxmox-api-go cloned_update_config