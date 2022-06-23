# On-site Demo Example

This Terraform configuration file provisions 5 instances as per the demo script:

> Next, they will provision their first 5 instances using Terraform: 3 web instances (2 cpus, 8gb DRAM, default 100GB disks) and 2 for DB (8 cpus, 32gb DRAM, 500GB disks) using an Oxide-provided OS image. The default VPC will be created automatically.

This configuration file does the following:

1. Retrieves data about the organization, project and global image. This configuration file assumes there will only be one of each and those are the ones we'll be using for the demo.
2. Using the data retrieved from the previous step, 5 disks are created; 3 100GiB disks for the web instances and 2 500GiB disks for the DB instances.
3. Finally, using data from the first and second steps, the 5 instances are created using the default VPC and subnet; 3 web instances with 2 CPUs and 8GiB of memory, and 2 DB instances with 8 CPUs and 32GiB of memory.

To try out this configuration file follow the [instructions](https://github.com/oxidecomputer/terraform-provider-oxide-demo/#using-the-provider) from the README.
