## Overview

The Black Duck OpsSight Connector provides software composition analysis of open-source components of containers in OpenShift v3.x and Kubernetes clusters. 

Pre-existing images and pods are automatically discovered, scanned and monitored. When a new image or pod is discovered, the OpsSight Connector kicks off a scan engine container to perform the scan and upload the results to your Black Duck instance.

## Getting involved 

We expect most upstream contributions to come to the blackducksoftware perceptor/perceiver/perceptor-scan/perceptor-protoform projects.

This project is under active development. We welcome all contributions, most of which will probably be to the aforementioned upstream repositories. 

## Release Status

See the releases tab for the latest OpsSight releases. If you identify an issue, please raise it, or better yet propose a solution.

## Developer notes

### Drafting a point release

How to draft a new OpsSight release, for the developers, using OpsSight 2.0.3 as an example:

- Tag branches in perceptor, perceiver, perceptor-scanner, perceptor-protoform ... which are changing (i.e. tag your perceptor branch you want to merge as release-2.0.3).
- Clone this repo
- Checkout the `release-2.0.x` opssight-connector branch
- Edit the opssight-connector `Gopkg.toml` file to point to the branch in the corresponding repo which was just updated.
- Run `dep ensure -update`
- Push your changes to the `release-2.0.x` opssight-connector branch.
- Now, the `build.properties` file should be something like 2.0.3-SNAPSHOT, and will automatically be updated once the downstream jenkins build is completed.

## License


Apache License 2.0

