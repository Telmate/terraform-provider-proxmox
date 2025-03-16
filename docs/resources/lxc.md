# LXC Resource

This resource creates and manages a Proxmox LXC container.

## Example Usage

### Basic example

```hcl
resource "proxmox_lxc" "basic" {
  target_node  = "pve"
  hostname     = "lxc-basic"
  ostemplate   = "local:vztmpl/ubuntu-20.04-standard_20.04-1_amd64.tar.gz"
  password     = "BasicLXCContainer"
  unprivileged = true

  // Terraform will crash without rootfs defined
  rootfs {
    storage = "local-zfs"
    size    = "8G"
  }

  network {
    name   = "eth0"
    bridge = "vmbr0"
    ip     = "dhcp"
  }
}
```

### Multiple mount points

-> By specifying `local-lvm:12` for the `mountpoint.storage` attribute in the first `mountpoint` block below, a volume
will be automatically created for the LXC container. For more information on this behaviour,
see [Storage Backed Mount Points](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_storage_backed_mount_points).

```hcl
resource "proxmox_lxc" "multiple_mountpoints" {
  target_node  = "pve"
  hostname     = "lxc-multiple-mountpoints"
  ostemplate   = "local:vztmpl/ubuntu-20.04-standard_20.04-1_amd64.tar.gz"
  unprivileged = true
  ostype       = "ubuntu"

  ssh_public_keys = <<-EOT
    ssh-rsa <public_key_1> user@example.com
    ssh-ed25519 <public_key_2> user@example.com
  EOT

  // Terraform will crash without rootfs defined
  rootfs {
    storage = "local-zfs"
    size    = "8G"
  }

  // Storage Backed Mount Point
  mountpoint {
    key     = "0"
    slot    = 0
    storage = "local-lvm"
    mp      = "/mnt/container/storage-backed-mount-point"
    size    = "12G"
  }

  // Bind Mount Point
  mountpoint {
    key     = "1"
    slot    = 1
    storage = "/srv/host/bind-mount-point"
    // Without 'volume' defined, Proxmox will try to create a volume with
    // the value of 'storage' + : + 'size' (without the trailing G) - e.g.
    // "/srv/host/bind-mount-point:256".
    // This behaviour looks to be caused by a bug in the provider.
    volume  = "/srv/host/bind-mount-point"
    mp      = "/mnt/container/bind-mount-point"
    size    = "256G"
  }

  // Device Mount Point
  mountpoint {
    key     = "2"
    slot    = 2
    storage = "/dev/sdg"
    volume  = "/dev/sdg"
    mp      = "/mnt/container/device-mount-point"
    size    = "32G"
  }

  network {
    name   = "eth0"
    bridge = "vmbr0"
    ip     = "dhcp"
    ip6    = "dhcp"
  }
}
```

### LXC with advanced features enabled

```hcl
resource "proxmox_lxc" "advanced_features" {
  target_node  = "pve"
  hostname     = "lxc-advanced-features"
  ostemplate   = "local:vztmpl/ubuntu-20.04-standard_20.04-1_amd64.tar.gz"
  unprivileged = true

  ssh_public_keys = <<-EOT
    ssh-rsa <public_key_1> user@example.com
    ssh-ed25519 <public_key_2> user@example.com
  EOT

  features {
    fuse    = true
    nesting = true
    mount   = "nfs;cifs"
  }

  // Terraform will crash without rootfs defined
  rootfs {
    storage = "local-zfs"
    size    = "8G"
  }

  // NFS share mounted on host
  mountpoint {
    slot    = "0"
    storage = "/mnt/host/nfs"
    mp      = "/mnt/container/nfs"
    size    = "250G"
  }

  network {
    name   = "eth0"
    bridge = "vmbr0"
    ip     = "10.0.0.2/24"
    ip6    = "auto"
  }
}
```

### Clone basic example

```hcl
resource "proxmox_lxc" "basic" {
  target_node = "pve"
  hostname    = "lxc-clone"
  #id of lxc container to clone
  clone       = "8001"
}
```

## Argument Reference

### Required

The following arguments must be defined when using this resource:

* `target_node` - A string containing the cluster node name.

### Optional

-> While the following arguments are optional, some have child arguments that are required when using the parent
argument (e.g. `name` in the `network` attribute). These child arguments have been marked with "__(required)__".

The following arguments may be optionally defined when using this resource:

* `ostemplate` - The [volume identifier](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_volumes) that points to
  the OS template or backup file.
* `arch` - Sets the container OS architecture type. Default is `"amd64"`.
* `bwlimit` - A number for setting the override I/O bandwidth limit (in KiB/s).
* `clone` - The lxc vmid to clone
* `clone_storage` - Target storage for full clone.
* `cmode` - Configures console mode. `"tty"` tries to open a connection to one of the available tty devices. `"console"`
  tries to attach to `/dev/console` instead. `"shell"` simply invokes a shell inside the container (no login). Default
  is `"tty"`.
* `console` - A boolean to attach a console device to the container. Default is `true`.
* `cores` - The number of cores assigned to the container. A container can use all available cores by default.
* `cpulimit` - A number to limit CPU usage by. Default is `0`.
* `cpuunits` - A number of the CPU weight that the container possesses. Default is `1024`.
* `description` - Sets the container description seen in the web interface.
* `features` - An object for allowing the container to access advanced features.
    * `fuse` - A boolean for enabling FUSE mounts.
    * `keyctl` - A boolean for enabling the `keyctl()` system call.
    * `mount` - Defines the filesystem types (separated by semicolons) that are allowed to be mounted.
    * `nesting` - A boolean to allow nested virtualization.
* `force` - A boolean that allows the overwriting of pre-existing containers.
* `full` - When cloning, create a full copy of all disks. This is always done when you clone a normal CT. For CT
  template it creates a linked clone by default.
* `hastate` - Requested HA state for the resource. One of "started", "stopped", "enabled", "disabled", or "ignored". See
  the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more
  info.
* `hagroup` - The HA group identifier the resource belongs to (requires `hastate` to be set!). See
  the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more
  info.
* `hookscript` - A string
  containing [a volume identifier to a script](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_hookscripts_2)
  that will be executed during various steps throughout the container's lifetime. The script must be an executable file.
* `hostname` - Specifies the host name of the container.
* `ignore_unpack_errors` - A boolean that determines if template extraction errors are ignored during container
  creation.
* `lock` - A string for locking or unlocking the VM.
* `memory` - A number containing the amount of RAM to assign to the container (in MB).
* `mountpoint` - An object for defining a volume to use as a container mount point. Can be specified multiple times.
    * `mp` __(required)__ - The path to the mount point as seen from inside the container. The path must not contain
      symlinks for security reasons.
    * `size` __(required)__ - Size of the underlying volume. Must end in T, G, M, or K (e.g. `"1T"`, `"1G"`, `"1024M"`
      , `"1048576K"`). Note that this is a read only value.
    * `slot` __(required)__ - A string containing the number that identifies the mount point (i.e. the `n`
      in [`mp[n]`](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#pct_mount_points)).
    * `key` __(required)__ - The number that identifies the mount point (i.e. the `n`
      in [`mp[n]`](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#pct_mount_points)).
    * `storage` __(required)__ - A string containing
      the [volume](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_storage_backed_mount_points)
      , [directory](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_bind_mount_points),
      or [device](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_device_mount_points) to be mounted into the
      container (at the path specified by `mp`). E.g. `local-lvm`, `local-zfs`, `local` etc.
    * `acl` - A boolean for enabling ACL support. Default is `false`.
    * `backup` - A boolean for including the mount point in backups. Default is `false`.
    * `quota` - A boolean for enabling user quotas inside the container for this mount point. Default is `false`.
    * `replicate` - A boolean for including this volume in a storage replica job. Default is `false`.
    * `shared` - A boolean for marking the volume as available on all nodes. Default is `false`.
* `nameserver` - The DNS server IP address used by the container. If neither `nameserver` nor `searchdomain` are
  specified, the values of the Proxmox host will be used by default.
* `network` - An object defining a network interface for the container. Can be specified multiple times.
    * `name` __(required)__ - The name of the network interface as seen from inside the container (e.g. `"eth0"`).
    * `bridge` - The bridge to attach the network interface to (e.g. `"vmbr0"`).
    * `firewall` - A boolean to enable the firewall on the network interface.
    * `gw` - The IPv4 address belonging to the network interface's default gateway.
    * `gw6` - The IPv6 address of the network interface's default gateway.
    * `hwaddr` - A string to set a common MAC address with the I/G (Individual/Group) bit not set. Automatically
      determined if not set.
    * `ip` - The IPv4 address of the network interface. Can be a static IPv4 address (in CIDR notation), `"dhcp"`,
      or `"manual"`.
    * `ip6` - The IPv6 address of the network interface. Can be a static IPv6 address (in CIDR notation), `"auto"`
      , `"dhcp"`, or `"manual"`.
    * `mtu` - A string to set the MTU on the network interface.
    * `rate` - A number that sets rate limiting on the network interface (Mbps).
    * `tag` - A number that specifies the VLAN tag of the network interface. Automatically determined if not set.
* `onboot` - A boolean that determines if the container will start on boot. Default is `false`.
* `ostype` - The operating system type, used by LXC to set up and configure the container. Automatically determined if
  not set.
* `password` - Sets the root password inside the container.
* `pool` - The name of the Proxmox resource pool to add this container to.
* `protection` - A boolean that enables the protection flag on this container. Stops the container and its disk from
  being removed/updated. Default is `false`.
* `restore` - A boolean to mark the container creation/update as a restore task.
* `rootfs` - An object for configuring the root mount point of the container. Can only be specified once.
    * `size` __(required)__ - Size of the underlying volume. Must end in T, G, M, or K (e.g. `"1T"`, `"1G"`, `"1024M"`
      , `"1048576K"`). Note that this is a read only value.
    * `storage` __(required)__ - A string containing
      the [volume](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_storage_backed_mount_points)
      , [directory](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_bind_mount_points),
      or [device](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#_device_mount_points) to be mounted into the
      container (at the path specified by `mp`). E.g. `local-lvm`, `local-zfs`, `local` etc.
* `searchdomain` - Sets the DNS search domains for the container. If neither `nameserver` nor `searchdomain` are
  specified, the values of the Proxmox host will be used by default.
* `ssh_public_keys` - Multi-line string of SSH public keys that will be added to the container. Can be defined
  using [heredoc syntax](https://www.terraform.io/docs/configuration/expressions/strings.html#heredoc-strings).
* `start` - A boolean that determines if the container is started after creation. Default is `false`.
* `startup` -
  The [startup and shutdown behaviour](https://pve.proxmox.com/pve-docs/pve-admin-guide.html#pct_startup_and_shutdown)
  of the container.
* `swap` - A number that sets the amount of swap memory available to the container. Default is `512`.
* `tags` - Tags of the container, semicolon-delimited (e.g. "terraform;test"). This is only meta information.
* `template` - A boolean that determines if this container is a template.
* `tty` - A number that specifies the TTYs available to the container. Default is `2`.
* `unique` - A boolean that determines if a unique random ethernet address is assigned to the container.
* `unprivileged` - A boolean that makes the container run as an unprivileged user. Default is `false`.
* `vmid` - A number that sets the VMID of the container. If set to `0`, the next available VMID is used. Default is `0`.
* `current_node` __(computed)__ - A string that shows on which node the LXC guest exists.|

## Attribute Reference

No additional attributes are exported by this resource.
