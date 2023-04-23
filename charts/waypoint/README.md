# Waypoint Helm Chart

This repository contains the official HashiCorp Helm chart for installing
and configuring Waypoint on Kubernetes.

This Helm chart only works with Waypoint 0.6+.

## Prerequisites

To use the charts here, [Helm](https://helm.sh/) must be configured for your
Kubernetes cluster. Setting up Kubernetes and Helm is outside the scope of
this README. Please refer to the Kubernetes and Helm documentation.

The versions required are:

  * **Helm 3.0+** - This is the earliest version of Helm tested. It is possible
    it works with earlier versions but this chart is untested for those versions.
  * **Kubernetes 1.20+** - This is the earliest version of Kubernetes tested.
    It is possible that this chart works with earlier versions but it is
    untested.

## Usage

#### Install the Helm Chart

```sh
$ helm repo add hashicorp https://helm.releases.hashicorp.com
"hashicorp" has been added to your repositories

$ helm install waypoint hashicorp/waypoint
...
```

#### Log In (Waypoint CLI)

From the same machine that ran the `helm install`, run the following
command to perform an initial login to the newly installed Waypoint server:

```sh
$ waypoint login -from-kubernetes
```

You can then run `waypoint ui` to open the web UI.

#### Log In (Manual)

If you do not want to use the Waypoint CLI, you can use `kubectl` or a
Kubernetes dashboard to find the Waypoint server address and token. First,
wait for the service to become ready:

```sh
# Not ready yet
$ kubectl get service waypoint-ui
NAME              TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                        AGE
waypoint-ui       LoadBalancer   10.245.138.80   <pending>     443:30836/TCP,9701:31734/TCP   4m42s

# Ready! (See External IP)
$ kubectl get service waypoint-ui
NAME          TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                        AGE
waypoint-ui   LoadBalancer   10.245.138.80   143.244.211.64   443:30836/TCP,9701:31734/TCP   9m15s
```

Copy the external IP. In this example it is `143.244.211.64`.

The Helm chart puts the initial token in the `waypoint-server-token`
Kubernetes Secret. Once you read this token, you can delete or rotate it
if you want.

Get the token:

```sh
$ kubectl get secret waypoint-server-token -o jsonpath="{.data.token}" | base64 --decode
3K4wQUdH1dfFHsFkYxmzzMHimJj6UE9tH4PhBQcLK9WACNmxHwmEJhqrvVGwVxaT2KUNtFUCAr7Wd3ci5NmFm6sRuKutzsn7CGs71ip2bnEywfEyxt7eaBvsn3kbCMjKzWPrEonR4Q7jgt6k⏎
```

You can now use this token to log in.
