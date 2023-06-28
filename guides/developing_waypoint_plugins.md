# Developing waypoint plugins

This guide talks through the basics of working with a (new) Waypoint plugin.

## Get a working plan / config in `executors`

First things first, you want to make sure you have a working, end to end plan in `executors`. Please refer to the [developing plans guide](developing_executor_plans.md) first. In short:

Run `workers-executors`:

```bash
$ nuonctl service run-local --name=workers-executors
```

Next, grab a valid install-id that has an active org. You can find this in stage.

Print a build config:
```bash
$ nuonctl builds --install-id=<install-id> --preset=helm_chart_private preset-waypoint-config
```

_Note_ your `build-id`, as it is a required input for printing a deploy config.

Print a sync-image config:
```bash
$ nuonctl deploys --sync=true --install-id=<install-id> --build-id=<build-id> waypoint-config
```

Print a deploy config:
```bash
$ nuonctl deploys --sync=false --install-id=<install-id> --build-id=<build-id> waypoint-config
```

## Publish "dev versions" of plugins

`workers-executors` is set up so you can publish development versions of your plugin for use while iterating. We have a dedicated `ecr` repo that the executors will read from when running by default in `dev`.

```bash
$ nctl scripts exec build-plugin waypoint-plugin-noop
```

**NOTE:** this backdoor always uses the most recent dev plugin. If someone else is debugging plugins at the same time, please communicate with them.

## Execute a plan

You can execute any plan by running `plan-and-execute`.

```bash
$ nuonctl deploys --sync=false --install-id=<install-id> --build-id=<build-id> plan-and-execute
```

## Debugging

There are three ways to debug jobs running:

### Connect to the waypoint server locally

You can connect to a waypoint server locally using the `nuonctl` script to automatically login:

```bash
$ nuonctl scripts exec waypoint-org-login <org-id>
```

From there, jobs are the most useful:

```bash
# print all jobs
$ waypoint job list -desc -json

# debug specific job
$ waypoint job inspect 01GRMEFXKANPNZ8WFEF70EQT9S
```


### Debug org namespace

There is a `nuonctl` script that will print out the useful information for an org during a run. Things like a bad image, container logs and more will show up here:

```bash
$ nctl scripts exec waypoint-odr-debug <org-id>
```

### Executors output / temporal

The executors themselves will output some errors locally. You can also access `localhost:8233` to see the temporal jobs as well. Generally speaking, all logs that are sent back to the `terminal.UI` type that waypoint provides the runtime will be visible here.

