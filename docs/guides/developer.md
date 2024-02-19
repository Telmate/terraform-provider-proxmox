# Overview
This guide aims to help developers who are new to this project so more people
can help investigate any problems they may run into.

# Compilation
Instructions on how to compile the provider, and cause terraform to use that
newly compiled executable, see the
[installation guide](https://github.com/Telmate/terraform-provider-proxmox/blob/new_developer_setup/docs/guides/installation.md#compile-the-executables-with-go).

You may want to specify the namespace and specific path to the plugin to make
sure terraform is getting the correct executable. If you are using the default
example domain as your namespace, it'd look like this:

```
terraform {
  required_providers {
    proxmox = {
      source = "registry.example.com/telmate/proxmox"
      #source = "Telmate/proxmox"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.14"
}
```

After changing that, you'll need to update and upgrade your terraform files:

```
terraform get -update
terraform init -upgrade
```

And finally, you'll want to check to make sure you only see the v1.0.0 entry
when you look at the providers that terraform reports about.

```
terraform version
```

If you are going to be copying different executables into that same location
repeatedly, you'll need to know that the hash of the executable is stored in
.terraform.lock.hcl. You will have to either manually remove the block for your
provisioner or just remove the file entirely before running the usual
terraform get/init commands listed above.

# Debugging
Instructions on how to enable debug logging are located
[here](https://registry.terraform.io/providers/Telmate/proxmox/latest/docs#pm_log_enable).

# Going deeper
Much of the code for the provider is not actually in this repo. It's in a
library repo called proxmox-api-go. When you build the provider, the build
system will check out a specific commit of that repo to get the code.

This is controlled by [go.sum](https://github.com/Telmate/terraform-provider-proxmox/blob/master/go.sum#L5-L6)

The convention seems to be `Version-Date-CommitHash`. As an example, the
following was commit `31826f2fdc39` that was checked in on 2023-12-07:

```
github.com/Telmate/proxmox-api-go v0.0.0-20231207182448-31826f2fdc39 h1:0MvktdAFWIcc9F4IwQls2Em1F9z2LUZR1fSVm1PkKfM=
github.com/Telmate/proxmox-api-go v0.0.0-20231207182448-31826f2fdc39/go.mod h1:xOwyTd8uC2IiYfmjwCVU2fTTVToFCm9yxJzn4cd7rPw=
```

If you want to make changes to the library (e.g. to add debug print
statements), you'll want to tell the Go compiler to look in a local directory
for your modified library instead of getting it from github. To do this, you
can add the following line to the end of go.mod:

```
replace "github.com/Telmate/proxmox-api-go" => "../proxmox-api-go"
```

This will let you experiment locally without having to push things up to github
and update hashes.
