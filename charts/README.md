# Charts

This directory contains both internal shared charts, as well as vendored external charts.

## Vendoring charts

We currently vendor the following charts:

* `waypoint` - Vendored from https://github.com/hashicorp/waypoint-helm and used to deploy waypoint servers and runners
* `temporal` - Vendored from https://github.com/temporalio/helm-charts and used to manage our temporal instances in
  stage and prod.

If you are updating a chart version, please check out the latest tag and copy the source files in here. We aim to make minimal changes here, to allow us to continue pulling upstream changes in.

## Why vendor charts?

We've long tried to avoid vendoring our own charts, and this has led to a few problems:

* forced us to manage custom k8s resources alongside helm charts (ie: in workers-orgs, we manually provision a public
  service).
* consistent installation - we can ensure that all charts are installed the same way.
