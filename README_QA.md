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

... Rob Rati to complete this section ... 
