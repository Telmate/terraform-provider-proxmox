# Storage Template Resource

This resource creates and manages CT Templates to create LXC containers.

## Example Usage

> [!Note]
> A list of available templates can be found by executing the following command on one of the nodes :
> ```sh
> pveam available
> ```

```hcl
resource "proxmox_storage_template" "template_example" {
  pve_node = "pve"
  storage = "local"
  template = "alpine-3.22-default_20250617_amd64.tar.xz"
}

resource "proxmox_lxc" "lxc_example" {
  ...
  ostemplate              = proxmox_storage_template.template_example.os_template
}
```
