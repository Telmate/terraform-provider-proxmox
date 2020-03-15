# Debug Errors

While applying your terraform setup, the Proxmox API can return errors. By increasing the verbosity of errors you can
debug more easily what is actually going. You achieve this by passing the environment variable `TF_LOG`.

With the following command, you can see all requests to the Proxmox.

```bash
TF_LOG=debug terraform apply
```
