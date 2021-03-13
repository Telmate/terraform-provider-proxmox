# Proxmox Provider

A Terraform provider is responsible for understanding API interactions and exposing resources. The Proxmox provider
uses the Proxmox API. This provider exposes two resources: [proxmox_vm_qemu](resources/vm_qemu.md) and [proxmox_lxc](resources/lxc.md).

## Creating the connection

When connecting to the Proxmox API, the provider has to know at least three parameters: the URL, username and password.
One can supply fields using the provider syntax in Terraform. It is recommended to pass secrets through environment 
variables.

```bash
export PM_PASS=password
```

```hcl
provider "proxmox" {
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
    pm_user = "terraform-user@pve"
}
```

## Argument Reference

The following arguments are supported in the provider block:

* `pm_api_url` - (Required; or use environment variable `PM_API_URL`) This is the target Proxmox API endpoint.
* `pm_user` - (Required; or use environment variable `PM_USER`) The user, maybe required to include @pam.
* `pm_password` - (Required; sensitive; or use environment variable `PM_PASS`) The password.
* `pm_otp` - (Optional; or use environment variable `PM_OTP`) The  2FA OTP code.
* `pm_tls_insecure` - (Optional) Disable TLS verification while connecting.
* `pm_parallel` - (Optional; defaults to 4) Allowed simultaneous Proxmox processes (e.g. creating resources).
* `pm_log_enable` - (Optional; defaults to false) Enable debug logging, see the section below for logging details.
* `pm_log_levels` - (Optional) A map of log sources and levels.
* `pm_log_file` - (Optional; defaults to "terraform-plugin-proxmox.log") If logging is enabled, the log file the provider will write logs to.
* `pm_timeout` - (Optional; defaults to 300) Timeout value (seconds) for proxmox API calls.

Additionally, one can set the `PM_OTP_PROMPT` environment variable to prompt for OTP 2FA code (if required).

## Logging

The provider is able to output detailed logs upon request. Note that this feature is intended for development purposes, but could also be used to help investigate bugs. For example: the following code when placed into the provider "proxmox" block will enable loging to the file "terraform-plugin-proxmox.log".  All log sources will default to the "debug" level, and any stdout/stderr from sublibraries (proxmox-api-go) will be silenced (set to non-empty string to enable).

```hcl
  pm_log_enable = true
  pm_log_file = "terraform-plugin-proxmox.log"
  pm_log_levels = {
    _default = "debug"
    _capturelog = ""
  }
```
