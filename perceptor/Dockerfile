FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME
ARG VERSION

# Container catalog requirements
COPY ./../LICENSE /licenses/
COPY ./core-help.1 /help.1

COPY ./perceptor ./perceptor

LABEL name="Black Duck OpsSight Core" \
      vendor="Black Duck Software" \
      release.version="$VERSION" \
      summary="Black Duck OpsSight Core" \
      description="This container is used to get images from opsight-image-processor and opssight-pod-processor in Kubernetes and Red Hat OpenShift clusters. It will send all those received images to the Black Duck scanner for scanning. It will also retrieve vulnerability and policy violations for each image periodically from the Black Duck Hub and update the scan results." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME"

CMD ["./perceptor"]
