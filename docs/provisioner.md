# Provisioner usage

Remove the temporary net1 adapter.
Inside the VM this usually triggers the routes back to the provisioning machine on net0.
```
  provisioner "proxmox" {
    action = "sshbackward"
  }

```

Replace the temporary net1 adapter with a new persistent net1.
```
  provisioner "proxmox" {
    action = "reconnect"
    net1 = "virtio,bridge=vmbr0,tag=99"
  }

```
If net1 needs a config other than DHCP you should prior to this use provisioner "remote-exec" to modify the network config.
