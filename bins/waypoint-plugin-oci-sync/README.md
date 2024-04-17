# Waypoint Plugin OCI Sync

This plugin handles syncing OCI images from the vendor's registry to the customer's registry, since we can't add that step into the waypoint-plugin-oci repo.

[Waypoint has an opinionated view on the minimum application lifecycle](https://developer.hashicorp.com/waypoint/docs/lifecycle). We can't have more than one push step, which means we can't add more than one Registry component to our OCI plugin. This is a problem because the first step of deploying an OCI from a vendor's registry to the a customer's install is syncing the OCI to the install's registry.

Since we can't add a second Registry component to waypoint-plugin-oci, we created this second plugin.
