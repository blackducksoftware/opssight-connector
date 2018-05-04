# OpsSight Connector Perceivers

## Building

### Building Binaries

To build the perceiver binaries, execute one of the two commands:

make
make local_build

The first option will build the binaries using a docker container so there is no need to a local golang installation whereas the second option is faster but requires a local golang installation.  Whichever option is used will result in the binaries being placed in an _output directory.

### Building Containers

Before building the perceiver containers, first set the REGISTRY environemnt varible.  This would be the registry where you would push the built containers.  For example if you wanted to push to a path in Google's Container Registry, you would set REGISTRY=gcr.io/pathtomyimages

Once the REGISTRY environment variable is set, run:

make container

Or to push the container straight to the registry run:

make push
