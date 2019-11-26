FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME
ARG VERSION

# Container catalog requirements
COPY ./LICENSE /licenses/
COPY ./help.1 /help.1

COPY ./opssight-artifactory-processor ./opssight-artifactory-processor

LABEL name="Black Duck OpsSight Artifactory Processor" \
      vendor="Black Duck Software" \
      release.version="$VERSION" \
      summary="Black Duck OpsSight Artifactory Processor" \
      description="This container is used to identify all images in an Artifactory instance. It will send all the identified Artifactory images to opssight-core for scanning. It will also retrieve vulnerability and policy violations for each image periodically from opssight-core and annotate and label the Artifactory image with the information." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME" \
      license="apache" \
      release="$VERSION" \
      version="$VERSION"

CMD ["./opssight-artifactory-processor"]
