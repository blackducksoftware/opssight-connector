# This is an introduction to opsite and perceptor for QA engineers.

The perceptor project is a distributed system for responding to events residing in
particular cloud native infrastructures.

The Perceivers are the cloud native extension points to the overall perceptor project, 
handling intercepting of events + responding to those events.

These opensource projects come together to make the opsite product.

# The opsite-connector product

The opsite product hardens the perceptor ecosystem into a 'downstream' artifact which
blackduck officially supports, with specific customizations that make it easy
for us to gaurantee performance and security specifications around.

# Getting started

NOTE: The first section is about learning kubernetes/openshift and understanding the upstream components of opsite are
not valuable to downstream testing.  

Once you understand the basics, skip to the next section, and focus on 'real QA' :).  

## First, understand the upstream and core components

The easiest way to get started is to use minikube, or minishift, to install the upstream 
perceptor components.
- First install minishift or minikube.
- git clone our installation tool, perceptor-protoform, and run a recipe in the install/ directory.  
- If you have issues file them in github.

It should be as simple as cloning perceptor-protoform, installing minishift or minikube,
and then running the install.sh scripts in the install/ directory from protofrom (either openshijft or kubernetes).

## Now, start testing downstream !


### Minishift OCP compatibility test:

- Create a red hat developer account

- Download the latest openshift-cdk https://developers.redhat.com/products/cdk/hello-world/ 

- Then enable OCP on it: 
`minishift config set vm-driver virtualbox ; minishift setup-cdk`

```
export MINISHIFT_USERNAME='<RED_HAT_USERNAME>'
export MINISHIFT_PASSWORD='<RED_HAT_PASSWORD>'
```

- Clone the connector, and run `git checkout release-2.0.x`.

- minishift ssh, then `sudo su`... At which point you can 'oc login', into localhost.  Use credentials 'admin'/'admin'.  Then it should allow you (since your root) to do:

`/var/lib/minishift/bin/oc  adm policy add-cluster-role-to-user cluster-admin developer`

Which effectively makes your developer account an administrative one.

- Run the installer: `./install.sh --hub-host 35.202.36.25 --hub-user sysadmin --hub-password blackduck --hub-port 443 --hub-max-concurrent-scans 2 -M -v 2.0.2-RC --pcp-namespace jasonia`.

