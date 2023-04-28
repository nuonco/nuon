# Waypoint Plugin Terraform

This plugin builds and applies terraform, using our internal terraform packaging + state backends.

## Local development

To run this plugin locally, you can log into a waypoint server (such as bootstrap) and use it directly:


Login to waypoint:
```bash
$ kubectx orgs-stage-main
$ waypoint login -vvv -from-kubernetes-namespace=waypoint -from-kubernetes -server-addr=waypoint.orgs-stage.nuon.co:9701 -server-tls-skip-verify
```

Build the plugin into your local bin:

```bash
$ go build -o ~/bins/waypoint-plugin-exp .
```

With a `waypoint.hcl` file such as the following, run `waypoint-up-local`:

```hcl
project = "my-project"

app "test" {
  build {
    use "terraform" {
      // TODO
    }
  }
}
```
