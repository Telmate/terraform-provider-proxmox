# Storage Template Resource

This resource creates and manages CT Templates to create LXC containers.

## Example Usage

> [!Warning]
> It is currently only possible to manage templates supported by Proxmox

The provider supports two ways to define a template. The first is to use the file
name from proxmox, and the second is to use the "package" and/or "version" fields
that can be found in the templates tab in the WebUI.

> [!Note]
> A list of available templates can be found by executing the following command on one of the nodes :
> ```sh
> pveam available
> ```

```hcl
resource "proxmox_storage_template" "template_example" {
  pve_node = "pve"
  storage = "local"
  template = {
    package = "alpine-3.22-default"
    # version = "20250617"  # Optional
    # file = "alpine-3.22-default_20250617_amd64.tar.xz"
  }
}

resource "proxmox_lxc" "lxc_example" {
  ...
  ostemplate              = proxmox_storage_template.template_example.os_template
}
```
