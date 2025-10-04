# Storage Template Resource

This resource creates and manages CT Templates to create LXC containers.

## Example Usage

> [!Note]
> A list of available templates can be found by executing the following command on one of the nodes :
> ```sh
> pveam available --section system
> ```

```hcl
resource "proxmox_storage_template" "example" {
  pve_node = "pve"
  storage = "local"
  template = "almalinux-9-default_20240911_amd64.tar.xz"
}
```
