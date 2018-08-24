## Overview

The opssight-connector provides integration between Black Duck Hub and OpenShift v3.x or Kubernetes. 

In the current implementation, pre-existing images and pods are automatically discovered and monitored. 

When a new image or pod is discovered, the integration kicks off a scan engine container to perform the scan and upload the results to your Black Duck Hub instance.

## Getting involved 

We expect most upstream contribubtions to come to the blackducksoftware perceptor/perceiver/perceptor-scan/perceptor-protoform projects.  This repo is the home for blackduck's downstream, supported product.

This project is under active development and has had no official releases. We welcome all contributions, most of which will 
probablby be to the aforementioned upstream repositories. 

## Release Status

Note that anyone attempting to use the code contained in here should expect rough edges and operational issues until release. If you identify an issue, please raise it, or better yet propose a solution.

## Developer notes

### Drafting a point release

Example of how to draft a new opssight release, for the developers, using 2.0.3 as an example.

- Tag branches in perceptor, perceiver, ... which are changing (i.e. tag your perceptor branch you want to merge as release-2.0.3).
- Clone this repo
- Checkout the release-2.0.x branch
- Edit the Gopkg.toml file to point to the corresponding branch your hotfixing.
- Run `dep ensure -update`
- Push your changes to the 2.0.x release opssight-connector branch.
- Now, the `build.properties` file should be something like 2.0.3-SNAPSHOT, and will automatically be updated once the downstream jenkins build is completed.

## License


Apache License 2.0

