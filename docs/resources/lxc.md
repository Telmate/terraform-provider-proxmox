# LXC Resource

Resources are the most important element in the Terraform language. Each resource block describes one or more 
infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

This resource manages a Proxmox LXC container.


```hcl
resource "proxmox_lxc" "lxc-test" {
    features {
        nesting = true
    }
    hostname = "terraform-new-container"
    network {
        name = "eth0"
        bridge = "vmbr0"
        ip = "dhcp"
        ip6 = "dhcp"
    }
    ostemplate = "shared:vztmpl/centos-7-default_20171212_amd64.tar.xz"
    password = "rootroot"
    pool = "terraform"
    storage = "local-lvm"
    target_node = "node-01"
    unprivileged = true
}
```
