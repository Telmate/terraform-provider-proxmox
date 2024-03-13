The files in this folder are here to help you troubleshoot issues by taking a
minimal example, which should work fine, and adding things from the terraform
file where you're having an issue.

These are also very helpful in reporting issues on the
[issue tracker](https://github.com/Telmate/terraform-provider-proxmox/issues)
so others can quickly reproduce your problem. When posting an issue, please
include the provider.tf file that you are using, and the smallest possible
issue.tf that recreates your problem. You don't need to include vars.tf, as
that's going to be different for everyone anyway.

In addition to providing terraform that reproduces the issue, please make sure
to mention what you are trying to accomplish, how you expected it to work, and
how it worked in practice. If you get any error messages, include those too.

Many times a debug log is not needed, but if you want to proactively create a
debug log and include that in your report, it may enable someone to answer your
question more quickly. To do so, uncomment the relevant lines of provider.tf and
include the .log file that is generated when you run `terraform apply`.
