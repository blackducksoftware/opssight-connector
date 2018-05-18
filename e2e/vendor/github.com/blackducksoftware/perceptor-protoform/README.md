# Perceptor-protoform: a cloud native administration utility for the perceptor ecosystem components:

- perceptor
- perceivers (openshift, kube, ...)
- perceptor-image-facade
- perceptor-scan

# Quick start

 - clone this repo

```
git clone git@github.com:blackducksoftware/perceptor-protoform.git
```

 - if running a Kubernetes cluster:
 
```
cd install/kube
```

 - if running an Openshift cluster: 
```
cd install/openshift
```

 - run the installer:
 
```
install.sh --prompt
```

## Tested Cluster Versions

Protoform has been run against the following clusters:

- Kubernetes 1.9.1
- Openshift Origin 3.6, 3.7 and 3.9

## Prerequisites

The user running the installation should be able to create service accounts with in-cluster API RBAC capabilities, and launch pods within them.  Specifically, protoform assumes access to oadm (for openshift users) or the ability to define RBAC objects (for kubernetes users).  

Protoform will attempt to detect your cluster type, and bootstrap all necessary components as needed.  This is done via environment variables, but the implementation is highly fluid right now, and we are leaning towards command line options once basic hardening of the core functionality is done.

## Run without cloning the source

Note that you can easily run without cloning, just pull down the `install` directory.
