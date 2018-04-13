FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME

# Container catalog requirements
COPY ./../LICENSE /licenses/
COPY ./core-help.1 /help.1

COPY ./perceptor ./perceptor

LABEL name="Black Duck OpsSight Core" \
      vendor="Black Duck Software" \
      release.version="1.0" \
      summary="Black Duck OpsSight Core" \
      description="This container is used to get all images from opsight-image-controller or opssight-pod-controller in any Kubernetes or Red Hat OpenShift cluster. It should send all those received images to blackduck scanner to scan them. It should also retrieve the vulnerability and policy violations of each image periodically from blackduck hub and update the scan results." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME"

CMD ["./perceptor"]
