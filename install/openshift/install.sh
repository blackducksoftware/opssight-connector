#!/bin/bash

unset DYLD_INSERT_LIBRARIES

echo "args = Namespace, Reg_key, version"

NS=$1
REG_KEY=$2
VERSION=$3

echo "Using the secret encoded in this file.  Change it before running, or press enter..."
read x

cat << EOF > /tmp/secret
apiVersion: v1
data:
  ADMIN_PASSWORD: YmxhY2tkdWNr
  POSTGRES_PASSWORD: YmxhY2tkdWNr
  USER_PASSWORD: YmxhY2tkdWNr
  HUB_PASSWORD: YmxhY2tkdWNr
kind: Secret
metadata:
  name: blackduck-secret
type: Opaque
EOF

oc new-project $NS

oc create -f /tmp/secret -n $NS

DOCKER_REGISTRY=docker.io
DOCKER_REPO=blackducksoftware

cat ../blackduck-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${TAG}/'${VERSION}'/g' | \
sed 's/${DOCKER_REGISTRY}/'$DOCKER_REGISTRY'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $DOCKER_REPO | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace=$NS -f -

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-np --type=NodePort --namespace=$NS

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-lb --type=LoadBalancer --namespace=$NS
