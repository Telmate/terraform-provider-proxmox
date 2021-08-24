# Proxmox Provider

A Terraform provider is responsible for understanding API interactions and exposing resources. The Proxmox provider uses the Proxmox API. This provider exposes two resources: [proxmox_vm_qemu](docs/resources/vm_qemu.md) and [proxmox_lxc](docs/resources/lxc.md).

## Creating the Proxmox user and role for terraform 

The particular priveledges required may change but here is a suitable starting point rather than using cluster-wide Administrator rights

Log into the Proxmox cluster or host using ssh (or mimic these in the GUI) then:
- Create a new role for the future terraform user.
- Create the user "terraform-prov@pve"
- Add the TERRAFORM-PROV role to the terraform-prov user

```bash
pveum role add TerraformProv -privs "VM.Allocate VM.Clone VM.Config.CDROM VM.Config.CPU VM.Config.Cloudinit VM.Config.Disk VM.Config.HWType VM.Config.Memory VM.Config.Network VM.Config.Options VM.Monitor VM.Audit VM.PowerMgmt Datastore.AllocateSpace Datastore.Audit"
pveum user add terraform-prov@pve --password <password>
pveum aclmod / -user terraform-prov@pve -role TerraformProv
```

The provider also supports using an API key rather than a password, see below for details. 

After the role is in use, if there is a need to modofy the privledges, simply issue the command showed, adding or removing priviledges as needed. 

```bash
pveum role modify TerraformProv -privs "VM.Allocate VM.Clone VM.Config.CDROM VM.Config.CPU VM.Config.Cloudinit VM.Config.Disk VM.Config.HWType VM.Config.Memory VM.Config.Network VM.Config.Options VM.Monitor VM.Audit VM.PowerMgmt Datastore.AllocateSpace Datastore.Audit"
```

For more information on existing roles and priviledges in Proxmox, refer to the vendor docs on [PVE User Management](https://pve.proxmox.com/wiki/User_Management)

## Creating the connection via username and password

When connecting to the Proxmox API, the provider has to know at least three parameters: the URL, username and password.
One can supply fields using the provider syntax in Terraform. It is recommended to pass secrets through environment
variables.

```bash
export PM_USER="terraform-user@pve"
export PM_PASS="password"
```

Note: these values can also be set in main.tf but users are encouraged to explore Vault as a way to remove secrets from their HCL.

```hcl
provider "proxmox" {
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
}
```

## Creating the connection via username and API token

```bash
export PM_API_TOKEN_ID="terraform-user@pve!mytoken"
export PM_API_TOKEN_SECRET="afcd8f45-acc1-4d0f-bb12-a70b0777ec11"
```

```hcl
provider "proxmox" {
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
}
```

## Argument Reference

The following arguments are supported in the provider block:

* `pm_api_url` - (Required; or use environment variable `PM_API_URL`) This is the target Proxmox API endpoint.
* `pm_user` - (Optional; or use environment variable `PM_USER`) The user, remember to include the authentication realm such as myuser@pam or myuser@pve.
* `pm_password` - (Optional; sensitive; or use environment variable `PM_PASS`) The password.
* `pm_api_token_id` - (Optional; or use environment variable `PM_API_TOKEN_ID`) This is an [API token](https://pve.proxmox.com/pve-docs/pveum-plain.html) you have previously created for a specific user.
* `pm_api_token_secret` - (Optional; or use environment variable `PM_API_TOKEN_SECRET`) This is a uuid that is only available when initially creating the token.
* `pm_otp` - (Optional; or use environment variable `PM_OTP`) The 2FA OTP code.
* `pm_tls_insecure` - (Optional) Disable TLS verification while connecting to the proxmox server.
* `pm_parallel` - (Optional; defaults to 4) Allowed simultaneous Proxmox processes (e.g. creating resources).
* `pm_log_enable` - (Optional; defaults to false) Enable debug logging, see the section below for logging details.
* `pm_log_levels` - (Optional) A map of log sources and levels.
* `pm_log_file` - (Optional; defaults to "terraform-plugin-proxmox.log") If logging is enabled, the log file the provider will write logs to.
* `pm_timeout` - (Optional; defaults to 300) Timeout value (seconds) for proxmox API calls.

Additionally, one can set the `PM_OTP_PROMPT` environment variable to prompt for OTP 2FA code (if required).

## Logging

The provider is able to output detailed logs upon request. Note that this feature is intended for development purposes, but could also be used to help investigate bugs. For example: the following code when placed into the provider "proxmox" block will enable loging to the file "terraform-plugin-proxmox.log".  All log sources will default to the "debug" level, and any stdout/stderr from sublibraries (proxmox-api-go) will be silenced (set to non-empty string to enable).

```hcl
provider "proxmox" {
  pm_log_enable = true
  pm_log_file = "terraform-plugin-proxmox.log"
  pm_log_levels = {
    _default = "debug"
    _capturelog = ""
  }
}
```
