# Local development
When developing this provider locally, you may find yourself confused how to actually use a freshly built `terraform-provider-proxmox`. This should come in handy if you're attempting to bootstrap development as quick as possible against a [Proxmox VE vagrant VM](https://github.com/rgl/proxmox-ve) for instance.

Most of this information was cobbled together from github issues such as [#25172](https://github.com/hashicorp/terraform/issues/25172).


# How to
This terraform configuration will apply to all terraform runs, so be aware of that. It allows you to use a locally built provider and avoid havint to futz with the [Terraform registry](https://www.terraform.io/docs/registry/api.html).
```
$ cat ~/.terraformrc
provider_installation {
  dev_overrides {}
  filesystem_mirror {
    path = "~/.terraform.d/plugins"
    include = ["registry.example.com/*/*"]
  }
  direct {}
}
```

Create some local terraform folders.
```
$ mkdir -p ~/.terraform.d/plugin-cache
$ mkdir -p ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/linux_amd64/
```

Your `terraform-provider-proxmox` binary will need to live inside the local registry folder.
```
$ ls -al ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/linux_amd64/
-rwxrwxr-x 1 phil phil 20352759 Feb 22 21:51 terraform-provider-proxmox
```

Change your directory over to where your proxmox states are located. Example [main.tf](./vagrant_example.tf) file.
```
$ pwd; ll
myproject
drwxrwxr-x 3 phil phil 4096 Feb 23 13:44 ./
drwxrwxr-x 5 phil phil 4096 Feb 22 21:56 ../
-rw-rw-r-- 1 phil phil 1045 Feb 23 13:43 main.tf
```

Verify that you can see your configured provider version
```
$ terraform providers
Providers required by configuration:
.
└── provider[registry.terraform.io/telmate/proxmox]
```

You'll probably need to do this each time you build a new copy of the terraform-provider-proxmox project.
```
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Reusing previous version of telmate/proxmox from the dependency lock file
- Using previously-installed telmate/proxmox

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

You should see plan output meaning that the provider can successfully authenticate with the proxmox vagrant.
```
$ terraform plan
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # proxmox_vm_qemu.phil will be created
  + resource "proxmox_vm_qemu" "phil" {
      + additional_wait        = 15
      + agent                  = 1
      + balloon                = 0
      + bios                   = "seabios"
```

Run a `terraform apply` and verify that your changes worked!
