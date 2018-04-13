FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME

# Container catalog requirements
COPY ./../LICENSE /licenses/
COPY ./deployer-help.1 /help.1

COPY ./protoform ./protoform

LABEL name="Black Duck OpsSight Deployer" \
      vendor="Black Duck Software" \
      release.version="1.0" \
      summary="Black Duck OpsSight Deployer" \
      description="This container is used to deploy all the containers that is required for Blackduck OpsSight product." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME"

CMD ["./protoform"]
