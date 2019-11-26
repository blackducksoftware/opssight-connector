FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME
ARG VERSION

# Container catalog requirements
COPY ./LICENSE /licenses/
COPY ./help.1 /help.1

COPY ./opssight-quay-processor ./opssight-quay-processor

LABEL name="Black Duck OpsSight Quay Processor" \
      vendor="Black Duck Software" \
      release.version="$VERSION" \
      summary="Black Duck OpsSight Quay Processor" \
      description="This container is used to identify all images in an Quay instance. It will send all the identified Quay images to opssight-core for scanning. It will also retrieve vulnerability and policy violations for each image periodically from opssight-core and annotate and label the Quay image with the information." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME" \
      license="apache" \
      release="$VERSION" \
      version="$VERSION"

CMD ["./opssight-quay-processor"]
