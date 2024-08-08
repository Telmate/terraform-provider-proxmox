# Proxmox Provider

A Terraform provider is responsible for understanding API interactions and exposing resources. The Proxmox provider uses
the Proxmox API. This provider exposes two resources: [proxmox_vm_qemu](resources/vm_qemu.md)
and [proxmox_lxc](resources/lxc.md).

## Creating the Proxmox user and role for terraform

The particular privileges required may change but here is a suitable starting point rather than using cluster-wide
Administrator rights

Log into the Proxmox cluster or host using ssh (or mimic these in the GUI) then:

- Create a new role for the future terraform user.
- Create the user "terraform-prov@pve"
- Add the TERRAFORM-PROV role to the terraform-prov user

```bash
pveum role add TerraformProv -privs "Datastore.AllocateSpace Datastore.AllocateTemplate Datastore.Audit Pool.Allocate Sys.Audit Sys.Console Sys.Modify VM.Allocate VM.Audit VM.Clone VM.Config.CDROM VM.Config.Cloudinit VM.Config.CPU VM.Config.Disk VM.Config.HWType VM.Config.Memory VM.Config.Network VM.Config.Options VM.Migrate VM.Monitor VM.PowerMgmt SDN.Use"
pveum user add terraform-prov@pve --password <password>
pveum aclmod / -user terraform-prov@pve -role TerraformProv
```

The provider also supports using an API key rather than a password, see below for details.

After the role is in use, if there is a need to modify the privileges, simply issue the command showed, adding or
removing privileges as needed.


Proxmox > 8:
```bash
pveum role modify TerraformProv -privs "Datastore.AllocateSpace Datastore.AllocateTemplate Datastore.Audit Pool.Allocate Sys.Audit Sys.Console Sys.Modify VM.Allocate VM.Audit VM.Clone VM.Config.CDROM VM.Config.Cloudinit VM.Config.CPU VM.Config.Disk VM.Config.HWType VM.Config.Memory VM.Config.Network VM.Config.Options VM.Migrate VM.Monitor VM.PowerMgmt SDN.Use"
```
Proxmox < 8:
```bash
pveum role modify TerraformProv -privs "Datastore.AllocateSpace Datastore.AllocateTemplate Datastore.Audit Pool.Allocate Sys.Audit Sys.Console Sys.Modify VM.Allocate VM.Audit VM.Clone VM.Config.CDROM VM.Config.Cloudinit VM.Config.CPU VM.Config.Disk VM.Config.HWType VM.Config.Memory VM.Config.Network VM.Config.Options VM.Migrate VM.Monitor VM.PowerMgmt"
```
For more information on existing roles and privileges in Proxmox, refer to the vendor docs
on [PVE User Management](https://pve.proxmox.com/wiki/User_Management)

## Creating the connection via username and password

When connecting to the Proxmox API, the provider has to know at least three parameters: the URL, username and password.
One can supply fields using the provider syntax in Terraform. It is recommended to pass secrets through environment
variables.

```bash
export PM_USER="terraform-prov@pve"
export PM_PASS="password"
```

Note: these values can also be set in main.tf but users are encouraged to explore Vault as a way to remove secrets from
their HCL.

```hcl
provider "proxmox" {
  pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
}
```

## Creating the connection via username and API token

```bash
export PM_API_TOKEN_ID="terraform-prov@pve!mytoken"
export PM_API_TOKEN_SECRET="afcd8f45-acc1-4d0f-bb12-a70b0777ec11"
```

```hcl
provider "proxmox" {
  pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
}
```

## Enable Debug Mode in proxmox-api-go

You can enable global debug mode for the provider underlying api client, using the new provider parameter. The default
setting is _false_

```hcl
provider "proxmox" {
  pm_debug = true
}
```

## Enable proxy server support

You can send all api calls from the provider api client to a proxy server rather than directly to proxmox itself. This
can make debugging easier. A nice proxy server is mitmproxy.

```hcl
provider "proxmox" {
  pm_proxy_server = "http://proxyurl:proxyport"
}
```

## Argument Reference

The following arguments are supported in the provider block:

- `pm_api_url` - (Required; or use environment variable `PM_API_URL`) This is the target Proxmox API endpoint.
- `pm_user` - (Optional; or use environment variable `PM_USER`) The user, remember to include the authentication realm
  such as myuser@pam or myuser@pve.
- `pm_password` - (Optional; sensitive; or use environment variable `PM_PASS`) The password.
- `pm_api_token_id` - (Optional; or use environment variable `PM_API_TOKEN_ID`) This is
  an [API token](https://pve.proxmox.com/pve-docs/pveum-plain.html) you have previously created for a specific user.
- `pm_api_token_secret` - (Optional; or use environment variable `PM_API_TOKEN_SECRET`) This uuid is only
  available when the token was initially created.
- `pm_otp` - (Optional; or use environment variable `PM_OTP`) The 2FA OTP code.
- `pm_tls_insecure` - (Optional) Disable TLS verification while connecting to the proxmox server.
- `pm_parallel` - (Optional; defaults to 4) Allowed simultaneous Proxmox processes (e.g. creating resources).
- `pm_log_enable` - (Optional; defaults to false) Enable debug logging, see the section below for logging details.
- `pm_log_levels` - (Optional) A map of log sources and levels.
- `pm_log_file` - (Optional; defaults to "terraform-plugin-proxmox.log") If logging is enabled, the log file the
  provider will write logs to.
- `pm_timeout` - (Optional; defaults to 300) Timeout value (seconds) for proxmox API calls.
- `pm_debug` - (Optional; defaults to false) Enable verbose output in proxmox-api-go
- `pm_proxy_server` - (Optional; defaults to nil) Send provider api call to a proxy server for easy debugging

Additionally, one can set the `PM_OTP_PROMPT` environment variable to prompt for OTP 2FA code (if required).

## Logging

The provider is able to output detailed logs upon request. Note that this feature is intended for development purposes,
but could also be used to help investigate bugs. For example: the following code when placed into the provider "proxmox"
block will enable loging to the file "terraform-plugin-proxmox.log". All log sources will default to the "debug" level.
To silence and any stdout/stderr from sub libraries (proxmox-api-go), remove or comment out `_capturelog`.

```hcl
provider "proxmox" {
  pm_log_enable = true
  pm_log_file   = "terraform-plugin-proxmox.log"
  pm_debug      = true
  pm_log_levels = {
    _default    = "debug"
    _capturelog = ""
  }
}
```
