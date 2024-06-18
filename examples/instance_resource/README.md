# Create basic instance

This example Terraform configuration file performs the following:

1. Creates a basic instance that starts up a basic ngnix web server.

To try it out make sure you have the following:

- A project.
- Previously set `OXIDE_HOST` and `OXIDE_TOKEN` environment variables

Alternatively, you can modify the configuration file with your own `project_id`. Although not recommended, if you do not wish to set the environment variables, you can use the `host` and `token` fields in the `oxide` provider block:

```hcl
provider "oxide" {
  host = "<your-host>"
  token = "<your-token>"
}
```

Once the instance is running and cloud-init has finished (you can check this via the serial console on the UI), copy the external IP from the output. On your browser's
navigation bar paste the external IP followed by port 80 e.g. 45.154.216.150:80.
You should get a basic HTML website with the text: "Heya 0xide!".