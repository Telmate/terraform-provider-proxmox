# How to get terraform to recognize third party provider

Third-party plugins (both providers and provisioners) can be manually installed into the user plugins directory, 
located at `%APPDATA%\terraform.d\plugins` on Windows and `~/.terraform.d/plugins` on other systems. Plugins come
with executables that have to be placed in the plugin directory.

## Compile the executables with Go

In order to build the required executables, [install Go](https://golang.org/doc/install) first. Then clone this
repository. Then run the following commands inside the cloned repository. Since this plugin is both a provider and 
provisioner in one, there are two install commands.

```
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provisioner-proxmox
```

Then create the executables. They are placed in the `bin` folder inside the repository.

```
make
```

## Copy executables to plugin directory

You need to copy these executables to the ~/.terraform.d directory which will also need to have a plugins directory 
created.

```shell
mkdir ~/.terraform.d/plugins
cp bin/terraform-provider-proxmox ~/.terraform.d/plugins
cp bin/terraform-provisioner-proxmox ~/.terraform.d/plugins
```

## Initialize Terraform

Now the plugin is installed, you can simply create a new terraform directory and do usual terraforming.

```
terraform init
```
