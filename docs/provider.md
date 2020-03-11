# Terraform Provider

A Terraform provider is responsible for understanding API interactions and exposing resources. The Proxmox provider
uses the Proxmox API. This provider exposes two resources: [proxmox_vm_qemu](resource_vm_qemu.md) and [proxmox_lxc](resource_lxc.md).

## Creating the connection

When connecting to the Proxmox API, the provider has to know at least three parameters: the URL, username and password.
One can supply fields using the provider syntax in Terraform. It is recommended to pass secrets through environment 
variables.

```
export PM_PASS=password
```

```tf
provider "proxmox" {
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
    pm_user = "terraform-user@pve"
}
```

## Argument Reference

The following arguments are supported in the provider block:

* `pm_api_url` - (Required; or use environment variable `PM_API_URL`) This is the target Proxmox API endpoint.
* `pm_user` - (Required; or use environment variable `PM_USER`) The user, maybe required to include @pam.
* `pm_password` - (Required; or use environment variable `PM_PASS`) The password.
* `pm_otp` - (Optional; or use environment variable `PM_OTP`) The  2FA OTP code.
* `pm_tls_insecure` - (Optional) Disable TLS verification while connecting.
* `pm_parallel` - (Optional; defaults to 4) Allowed simultaneous Proxmox processes (e.g. creating resources).

Additionally, one can set the `PM_OTP_PROMPT` environment variable to prompt for OTP 2FA code (if required).
