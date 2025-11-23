# LXC Guest Resource

This resource manages a LXC container.

To get up and running this is a minimal example:

```hcl
resource "proxmox_lxc_guest" "minimal-example" {
    name         = "minimal-example"
    power_state  = "running"
    node         = "pve-1"
    unprivileged = true
    password     = "yourpassword"
    template {
        file    = "alpine-3.22-default_20250617_amd64.tar.xz"
        storage = "local"
    }
    cpu {
        cores = 1
    }
    memory = 1024
    swap   = 512
    pool   = "my-pool"
    root_mount {
        size    = "4G"
        storage = "local-lvm"
    }
    network {
        id = 0
        name = "eth0"
        bridge = "vmbr0"
        ipv4_address = "192.168.1.100/24"
        ipv4_gateway = "192.168.1.1"
    }
}
```

## Argument reference

| Argument            | Type    | Default Value            | Description |
|:--------------------|---------|--------------------------|:------------|
| `clone`             | `nested`|                          | **Forces Recreation**: Clone configuration, see [Clone Reference](#clone-reference).|
| `cpu_architecture`  | `string`|                          | **Computed**: The CPU architecture.|
| `cpu`               | `nested`|                          | CPU configuration, see [CPU Reference](#cpu-reference).|
| `description`       | `string`| `"Managed by Terraform."`| Description of the guest container.|
| `dns`               | `nested`|                          | DNS configuration, see [DNS Reference](#dns-reference).|
| `guest_id`          | `int`   |                          | **Forces Recreation**, **Computed**: The numeric ID of the guest container also known as `vmid`. If not specified, an ID will be automatically assigned.|
| `memory`            | `int`   | `512`                    | The amount of memory to allocate to the guest in Megabytes.|
| `mount`             | `array` |                          | Storage mounts configured as individual array items, see [Mount Reference](#mount-reference).|
| `mounts`            | `nested`|                          | Storage mounts configured as nested sub items, see [Mounts Reference](#mounts-reference).|
| `name`              | `string`|                          | **Required**: The name of the container.|
| `network`           | `array` |                          | Network interfaces configured as individual array items, see [Network Reference](#network-reference).|
| `networks`          | `nested`|                          | Network interfaces configured as nested sub items, see [Networks Reference](#networks-reference).|
| `os`                | `string`|                          | **Computed**: The name of the OS inside the guest.|
| `password`          | `string`|                          | **Forces Recreation**, **Sensitive**: The password of the root user inside the guest container.|
| `pool`              | `string`|                          | The name of the pool the guest container should be a member of.|
| `power_state`       | `string`| `"running"`              | Power state of the guest, can be `"running"` or `"stopped"`.|
| `privileged`        | `bool`  |                          | **Forces Recreation**: If the guest is privileged or unprivileged. Can only be `true` or unset. Mutually exclusive with `unprivileged`.|
| `root_mount`        | `nested`|                          | **Required**: Configuration of the root/boot mount/disk of the guest container. **Note:** Size can only be increased, not decreased.|
| `ssh_public_key`    | `string`|                          | **Forces Recreation** SSH public key of the root user inside the guest container.|
| `start_at_node_boot`| `bool`  | `false`                  | Whether the guest should start automatically when the Proxmox node boots.|
| `startup_shutdown`  | `nested`|                          | Startup and shutdown configuration of the guest, see [Startup and Shutdown Reference](#startup-and-shutdown-reference).|
| `swap`              | `int`   | `512`                    | Amount of virtual memory of the guest that will b mapped to swap space on the PVE node.|
| `tags`              | `list`  | `[]`                     | List of tags to assign to the guest container.|
| `target_node`       | `string`|                          | Single node the guest should be on. If the guest is on a different node it will be migrated to this one.|
| `target_nodes`      | `array` |                          | List of nodes the guest should be on. If the guest is not on one of these nodes it will be migrated to one of them.|
| `unprivileged`      | `bool`  |                          | **Forces Recreation**: If the guest is unprivileged or privileged. Can only be `true` or unset. Mutually exclusive with `privileged`.|

### Clone Reference

The `clone` field is used to configure the clone settings. It may ony be specified once.

| Argument       | Type    | Default Value | Description |
|:---------------|---------|---------------|:------------|
| `id`           | `int`   |               | **Forces Recreation**: The numeric ID of the source container to clone from.|
| `linked`       | `bool`  | `false`       | **Forces Recreation**: Wheter the clone should be a linked clone.|
| `name`         | `string`|               | **Forces Recreation**: The name of the source container to clone from. Either `id` or `name` must be specified.|

### CPU Reference

The `cpu` field is used to configure the CPU settings. It may ony be specified once.

| Argument | Type | Default Value | Description |
|:---------|------|---------------|:------------|
| `cores`  | `int`| `0`           | Number of CPU cores of the guest, `0` means unlimited.|
| `limit`  | `int`| `0`           | CPU limit of the guests CPU cores, `0` means unlimited.|
| `units`  | `int`| `100`         | CPU units of the guest.|

### DNS Reference

The `dns` field is used to configure the DNS settings. It may ony be specified once.

| Argument      | Type    | Default Value | Description |
|:--------------|---------|---------------|:------------|
| `searchdomain`| `string`| `""`          | DNS searchdomain of the guest, inherits the PVE node config when empty.|
| `nameserver`  | `array` | `[]`          | DNS nameserver of the guest, inherits the PVE node config when empty.|

### Mount Reference

The `mount` field is used to configure the mount settings. It may be specified multiple times, each instance requires a unique `slot` value. `mount` is mutually exclusive with `mounts`.

| Argument          | Type    | Default Value | Description |
|:------------------|---------|---------------|:------------|
| `acl`             | `string`| `default`     | Mount acl configuration, can be one of `"true"`, `"false"`, `"default"`. Requires `type` = `data`.|
| `backup`          | `bool`  | `true`        | Wheter the mount will be included in backup tasks.|
| `guest_path`      | `string`|               | **Required**: Absolute path of the mount point inside the container guest, example: `"/mnt/data-mount`.|
| `host_path`       | `string`|               | **Required when `type` = `bind`**: Absolute path of the mount point on the PVE host, example: `"/mnt/pve-storage/data-mount`.|
| `option_discard`  | `bool`  | `true`        | Enable discart.|
| `option_lazy_time`| `bool`  | `true`        | Enable lazy time.|
| `option_no_atime` | `bool`  | `true`        | Enable no atime.|
| `option_no_device`| `bool`  | `true`        | Enable no device.|
| `option_no_exec`  | `bool`  | `true`        | Enable no exec.|
| `option_no_suid`  | `bool`  | `true`        | Enable no suid.|
| `quota`           | `bool`  | `false`       | Wheter data quota should be enabled. Requires top level `privileged` = `true`. Requires `type` = `data`.|
| `read_only`       | `bool`  | `false`       | Wheter the mount point is read only.|
| `replicate`       | `bool`  | `false`       | Wheter replication is enabled on the mount point.|
| `size`            | `string`|               | **Required when `type` = `data`**: Size of the mount.|
| `slot`            | `string`|               | **required**: The unique slot id of the mount. Must be prefixed with `mp`, example: `mp0`. Maximum amount of mounts is 256.|
| `storage`         | `string`|               | **Required when `type` = `data`**: Storage of the mount.|
| `type`            | `string`| `"data"`      | The type of mount point. Use `"bind"` for a bind mount to the PVE host and  `"data"` for a normal data disk.|

### Mounts Reference

The `mounts` field is used to configure the mount settings. It may only be specified once. `mounts` is mutually exclusive with `mount`. `mounts` has 256 sub items, with each sub item representing a unique slot, example: `mp0`, `mp1`, etc. Every slot has the following configuration options:

| Argument | Type    | Default Value | Description |
|:---------|---------|---------------|:------------|
| `bind`   | `nested`|               | Bind mount configuration, see [Bind Mounts Reference](#bind-mounts-reference).|
| `data`   | `nested`|               | Data mount configuration, see [Data Mounts Reference](#data-mounts-reference).|

#### Bind Mounts Reference

The `bind` field is used to configure a bind moun, And is mutually exclusive with `data`.

| Argument          | Type    | Default Value | Description |
|:------------------|---------|---------------|:------------|
| `guest_path`      | `string`|               | **Required**: Absolute path of the mount point inside the container guest, example: `"/mnt/data-mount`.|
| `host_path`       | `string`|               | **Required**: Absolute path of the mount point on the PVE host, example: `"/mnt/pve-storage/data-mount`.|
| `option_discard`  | `bool`  | `true`        | Enable discart.|
| `option_lazy_time`| `bool`  | `true`        | Enable lazy time.|
| `option_no_atime` | `bool`  | `true`        | Enable no atime.|
| `option_no_device`| `bool`  | `true`        | Enable no device.|
| `option_no_exec`  | `bool`  | `true`        | Enable no exec.|
| `option_no_suid`  | `bool`  | `true`        | Enable no suid.|
| `read_only`       | `bool`  | `false`       | Wheter the mount point is read only.|
| `replicate`       | `bool`  | `false`       | Wheter replication is enabled on the mount point.|

#### Data Mounts Reference

The `data` field is used to configure a data mount, And is mutually exclusive with `bind`.

| Argument          | Type    | Default Value | Description |
|:------------------|---------|---------------|:------------|
| `acl`             | `string`| `default`     | Mount acl configuration, can be one of `"true"`, `"false"`, `"default"`. Requires top level `privileged` = `true`.|
| `backup`          | `bool`  | `true`        | Wheter the mount will be included in backup tasks.|
| `guest_path`      | `string`|               | **Required**: Absolute path of the mount point inside the container guest, example: `"/mnt/data-mount`.|
| `option_discard`  | `bool`  | `true`        | Enable discart.|
| `option_lazy_time`| `bool`  | `true`        | Enable lazy time.|
| `option_no_atime` | `bool`  | `true`        | Enable no atime.|
| `option_no_device`| `bool`  | `true`        | Enable no device.|
| `option_no_exec`  | `bool`  | `true`        | Enable no exec.|
| `option_no_suid`  | `bool`  | `true`        | Enable no suid.|
| `quota`           | `bool`  | `false`       | Wheter data quota should be enabled. Requires top level `privileged` = `true`.|
| `read_only`       | `bool`  | `false`       | Wheter the mount point is read only.|
| `replicate`       | `bool`  | `false`       | Wheter replication is enabled on the mount point.|
| `size`            | `string`|               | **Required**: Size of the mount.|
| `storage`         | `string`|               | **Required**: Storage of the mount.|

### Network Reference

The `network` field is used to configure the network interfaces. It may be specified multiple times, each instance requires a unique `id` and `name` value. `network` is mutually exclusive with `networks`.

| Argument        | Type    | Default Value | Description |
|:----------------|---------|---------------|:------------|
| `bridge`        | `string`|               | **Required**: Bridge the network interface will be connected to.|
| `connected`     | `bool`  | `true`        | Wheter the network interface will be connected.|
| `firewall`      | `bool`  | `false`       | Wheter the network interface will be protected by the firewall.|
| `id`            | `string`|               | **Required**: The unique id of the network interface. Must be prefixed with `net`, example: `net0`. Maximum amount of network interfaces is 16.|
| `ipv4_address`  | `string`|               | IPv4 address of the network interface.|
| `ipv4_dhcp`     | `bool`  | `false`       | Wheter IPv4 DHCP is enabled on the network interface.|
| `ipv4_gateway`  | `string`|               | IPv4 gateway of the network interface.|
| `ipv6_address`  | `string`|               | IPv6 address of the network interface.|
| `ipv6_dhcp`     | `bool`  | `false`       | Wheter IPv6 DHCP is enabled on the network interface.|
| `ipv6_gateway`  | `string`|               | IPv6 gateway of the network interface.|
| `mac`           | `string`|               | MAC address of the network interface.|
| `mtu`           | `int`   |               | MTU of the network interface.|
| `name`          | `string`|               | **Required**: Name of the network interface inside the guest. The name must be unique example: `eth0`.|
| `rate_limit`    | `int`   |               | Rate limit of the network interface in Kbit/s. `0` means unlimited.|
| `slaac`         | `bool`  | `false`       | Wheter SLAAC is enabled on the network interface. Conflicts with IPv6 settings.|
| `vlan_native`   | `int`   |               | Native VLAN of the network interface.|

### Networks Reference

The `networks` field is used to configure the network interfaces. It may only be specified once. `networks` is mutually exclusive with `network`. `mounts` has 16 sub items, with each sub item representing a unique id, example: `net0`, `net1`, etc. Every slot has the following configuration options:

| Argument     | Type    | Default Value | Description |
|:-------------|---------|---------------|:------------|
| `bridge`     | `string`|               | **Required**: Bridge the network interface will be connected to.|
| `connected`  | `bool`  | `true`        | Wheter the network interface will be connected.|
| `firewall`   | `bool`  | `false`       | Wheter the network interface will be protected by the firewall.|
| `ipv4`       | `nested`|               | IPv4 configuration, see [IPv4 Reference](#ipv4-reference).|
| `ipv6`       | `nested`|               | IPv6 configuration, see [IPv6 Reference](#ipv6-reference).|
| `mac`        | `string`|               | MAC address of the network interface.|
| `mtu`        | `int`   |               | MTU of the network interface.|
| `name`       | `string`|               | **Required**: Name of the network interface inside the guest. The name must be unique example: `eth0`.|
| `rate_limit` | `int`   |               | Rate limit of the network interface in Kbit/s. `0` means unlimited.|
| `vlan_native`| `int`   |               | Native VLAN of the network interface.|

#### IPv4 Reference

| Argument     | Type    | Default Value | Description |
|:-------------|---------|---------------|:------------|
| `address`    | `string`|               | IPv4 address of the network interface.|
| `gateway`    | `string`|               | IPv4 gateway of the network interface.|
| `dhcp`       | `bool`  | `false`       | Wheter IPv4 DHCP is enabled on the network interface. Mutually exclusive with `address` and `gateway`|

#### IPv6 Reference

| Argument     | Type    | Default Value | Description |
|:-------------|---------|---------------|:------------|
| `address`    | `string`|               | IPv6 address of the network interface.|
| `gateway`    | `string`|               | IPv6 gateway of the network interface.|
| `dhcp`       | `bool`  | `false`       | Wheter IPv6 DHCP is enabled on the network interface. Mutually exclusive with `address`, `gateway` and `slaac`.|
| `slaac`      | `bool`  | `false`       | Wheter SLAAC is enabled on the network interface. Conflicts with IPv6 settings. Mutually exclusive with `address`, `gateway` and `dhcp`.|

### Startup and Shutdown Reference

The `startup_shutdown` field is used to configure the startup and shutdown settings. It may ony be specified once.

| Argument            | Type | Default Value | Description |
|:--------------------|------|---------------|:------------|
| `order`             | `int`| `-1`          | Startup order `-1` means any.|
| `shutdown_timeout`  | `int`| `-1`          | Shutdown timeout in seconds, `-1` means default.|
| `startup_delay`     | `int`| `-1`          | Startup delay in seconds, `-1` means default.|
