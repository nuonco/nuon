# Installer and Bundles

This RFC proposes a new concept - the Nuon installer. The installer is a CLI that can ship alongside any install and 
either give customer controls for a managed install, or let a customer create a bundle install that requires manually 
triggering and verifying the components.

The installer is a CLI that runs a webapp. It is customizable and manages state via the installer-api. The installer API 
is a blob-storage based API that can talk to the control-plane for managed installs, or keep all state locally. For 
bundle installs, the runner jobs will be executed directly via the installer and the user presented with an interface 
for managing workflows. For managed installs, the installer will show a resource view, healthcheck, policy and workflow 
view and expose approvals.

The installer's UI is just embedded in the CLI. It can be run locally to manage permissions, or can be run on a server 
with an OAUTH HTTP API.

This RFC presents some new concepts (bundles) and a way to have both managed installs (ie: cloudformation) installs and 
bundle installs managed via one workflow. It introduces primitives for custom components and resources.

# Concepts

## Bundles

Bundles turn an app-branch-run into a signal artifact. For an app branch run, if any install in the app is a "bundle" 
install, the app will then create a bundle. A bundle is:

1. a tar.gz bundle or OCI artifact with the contents of every component.
2. comes with a checksum / digest and hash.
3. can be downloaded manually and handed to someone, or downloaded automatically via the installer.

## Bundle Installs

Bundle installs are installs that are managed manually via the installer. The installer is run via the customer for 
pulling in updates, and then the installer will automatically unpack the bundle's contents, let the customer approve or 
deny them, and the trigger them.

The way to think about bundles are that they are 1:1 with an app-branch-run -- each app branch run will make a bundle 
behind the scenes. 

Each bundle will also encode a set of workflows that can be run. Runbooks, actions, and the standard workflows can be 
encoded in a format that goe into the installs. This means, all workflows have a 1:1 mapping into the bundle.

Installs should be able to define a private-key that the bundle is signed and verified against for two way encryption. 
then, the install can take that public key and use it to verify the full bundle?

## Installer

This replaces the "stack", an installer can run components either with the customer's current. The installer can execute 
a workflow from the local control-plane (ie: a bundle install), where it fetches a workflow that is marked as 
"execution-mode: installer".

The installer then has an approval view of changes, where the customer can have a view of the resources.

The installer will have a data model in the API. The installer will come online, and use HMAC or something else to auth 
with the control-plane. The installer will then be able to pull information down and store state for the installer 
including: app-id, org-id, install-id and the bucket details to connect too. The installer has to be able to make a 
bucket connection. Then, it will connect to that, and when we run the installer `nuon-installer run`, the installer will 
start, connect to the bucket and then start the API.

Permissions have to go through installer, and let the customer choose things. When bundles are used, you should be able 
to verify the bundle.

The installer should be customizable. The vendor should be able to style it, change it, etc. The web page and what not 
that it uses should all have a style sheet that is shipped with the installer. Make a config file (ie: a json or yaml) 
that has all of the config options for the installer. This would then bundle with it.

### Stack Semantic

Right now, the stack is something that will automatically let the customer spin up the Nuon runner and configure it. 
There's a few issues:
1. if it's a terraform stack, we have to manage a bunch of terraform state and can't handle permissiosn.
2. we need a story for components that can be executed via the installer. This means, any component can be an "installer 
   component" where it's only installable via the installer. Even a managed BYOC install will move the components into 
   the installer for the customer to run.
3. the state needs to be stored locally, and needs and s3 bucket for it.

### Installer Semantics

The installer is a CLI. it can be downloaded and connected to an install. Then, it will create an s3 bucket for storage.

### Bundle Management

The UI should make installs that are bundle managed first class. you should be able to see bundles via the UI. Bundles 
should be an app field -- bundle_enabled: true, and then that is used to control the view.

An install that is bundle based will need an entirely new view. We will need workflow statuses, component health checks 
and bundle statuses pushed to the api via the installer. So a limited set of information.

### Bundle Install Management

The installer is a CLI that embeds a UI. The UI then has an installer API that defines the thing it can do.

## Workflows

We are going to expand workflows to run in multiple contexts. Workflows can run either in the installer, be approved via 
the installer, or approved via the vendor.

# Execution

## Installer

The installer is a CLI that can do the following:
1. show a webview of all resources for an install. This can be powered via the control-plane API with a short-lived 
   token and fetch state from there or the local api.
2. take a bundle and verify it, and then execute it.
3. show approvals of an install, so a customer can approve them.

## Installer API 

The installer API is an API that powers the installer. The installer-api has one of two modes:

1. local mode - it reads and writes everything from the local s3 bucket. Resources, errors, other things.
2. remote mode - it talks to a remote control plane, and uses that to send data back and forth.

The purpose of the installer-api is to give us a clean way to show resources, health-checks, and more by exposing data 
that is stored locally in s3 (limited view) or full view from the control-plane api. In some ways, it's a backend for 
frontend, but is pluggable and can use a limited scope token or a full bucket (depending if it's bundle mode).

## Customer Managed Resources

Any component (or the sandbox) can be declared as a "customer-managed: true|false" where a customer can opt to manage 
the resource. If the customer manages the resource, then the installer will show the customer the underlying code as the 
"template" and can render things like the pulumi / terraform outputs + resources by parsing the code.

Then, the customer can pick and choose to create the resource. There are three classes of resources:

1. stack resources - this spins up the runner, the vpc, roles etc.
2. sandbox resources - can connect a cluster.
3. component resources - any resource should be able to be customer managed where a customer can create the resources 
   and push outputs or let the runner control it.

## Managed Install Controls

There are several classes of controls that the installer has for managed installs:

1. it can add / remove permissions - enable roles, etc.
2. it can enable/disable the runner.
3. it can approve / deny the install updates.

## Customization

The installer needs to be customizable. This means that it's OSS and the customer can then build their own distribution 
and/or connect it to different apps and things. to start, it should only work for a single install, but it can create 
installs later. The most important thing to suss out now is that it can manage the install and resources etc.
