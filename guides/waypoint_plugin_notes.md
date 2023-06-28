# Notes on waypoint plugins

## Build and Push

The distinction between `build` and `push` can be tricky. From what we've gathered thus far it works something like this:

* if `BuildODR` is defined, it will take preference over `build`.
* `Build` can _not_ accept an `AccessInfo` type, but `BuildODR` can. This is a good distinction because it's hard to
  persist state between a build in the ODR because it's using something like kaniko
* `Push` is completely optional - only used if you are not using an ODR build function

## Plugin process state

It seems that each plugin is initialized in a separate process. For instance, you can _not_ share in-process state between build and push.

## Plugin initialization

You can find the initialization code for picking plugins [here]().

Generally the way it works:

* the plugin runtime looks for a plugin named `waypoint-plugin-<name>`
* if that can't be found, it checks for `builtin`

Once the plugin is initialized, the plugin runtime does the following:

* initializes the plugin, by calling `AccessInfo`
* checks for `BuildODR` - if it exists calls it with or with out `AccessInfo`, based on the signature.
* calls `Push`

It's worth noting that if `BuildODR` does the push (such as the builtin docker ODR build), then it's up to the plugin to tell the Push function to not push it.

## How do plugin calls work?

The plugin runtime uses a dynamic reflection approach to identify the plugins to call.

Essentially, it loads protos from the waypoint plugin for common things:

* AccessInfo
* BuildOutput
* Artifact

From there, it uses https://github.com/hashicorp/go-argmapper to dynamically call a method on it.

## Mapping protos

Waypoint allows you to use protos between plugins. A good example of this is using the `docker.Image` proto that is the response of the `docker` build plugin as an input elsewhere.

The waypoint maps these plugins under the hood is using actual proto reflection. It doesn't just match the fields for equality, it actually needs to be able to ensure they are the same type.

## Entrypoint is unneeded for us

The entrypoint causes a ton of complexity:

* makes the docker build plugins harder - they actually have to inject an image
* has to run inside a real container
* requires a long lived connection to a server, outside the plugin
* requires a lot of infrastructure (ie: horizon)

We don't need this and honestly would prefer to not introduce this complexity for our customers. Instead, we plan on using plugins to do things like surface logs instead. This means we can:

* remove the entrypoint from our docs / requiring custom integrations for customers (ie: helm)
* remove having to run waypoint-horizon and horizon
* re-leverage our existing tools for plugins
