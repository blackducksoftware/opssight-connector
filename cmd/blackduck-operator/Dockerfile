FROM scratch

MAINTAINER Black Duck OpsSight Team

ARG LASTCOMMIT
ARG BUILDTIME
ARG VERSION

# Container catalog requirements
COPY ./LICENSE /licenses/
COPY ./help.1 /help.1

COPY ./blackduck-operator ./blackduck-operator

LABEL name="Black Duck Operator" \
      vendor="Black Duck Software" \
      release.version="$VERSION" \
      summary="Black Duck Operator" \
      description="This container is used to deploy the Black Duck Operators." \
      lastcommit="$LASTCOMMIT" \
      buildtime="$BUILDTIME" \
      license="apache" \
      release="$VERSION" \
      version="$VERSION"

CMD ["./blackduck-operator"]
