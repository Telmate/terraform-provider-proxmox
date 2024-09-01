# Pool Resource

This resource creates and manages a VM / LXC pool.

## Example Usage

```hcl
resource "proxmox_pool" "example" {
  poolid  = "example-pool" 
  comment = "Example of a pool"
}
```
