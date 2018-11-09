#!/bin/bash
echo "args = Namespace"

NS=$1
DOCKER_CONFIG_JSON_PATH="$2"
DOCKER_REGISTRY=$3
DOCKER_REPO=$4
VERSION=$5

if [[ ! -z "$3" ]]; then
  DOCKER_REGISTRY="$3"
else
  DOCKER_REGISTRY="registry.connect.redhat.com"
fi 

if [[ ! -z "$4" ]]; then
  DOCKER_REPO="$4"
else
  DOCKER_REPO="blackducksoftware"
fi

if [[ ! -z "$5" ]]; then
  VERSION="$5"
else
  VERSION="latest"
fi

cat ./opssight_with_min_parameter.yaml | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${TAG}/'${VERSION}'/g' | \
sed 's/${DOCKER_REGISTRY}/'$DOCKER_REGISTRY'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $DOCKER_REPO | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create -f -

sleep 20;

if [[ ! -z "$DOCKER_CONFIG_JSON_PATH" ]]; then
  oc project "$NS"
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$DOCKER_CONFIG_JSON_PATH" --type=kubernetes.io/dockerconfigjson
  oc secrets link default custom-registry-pull-secret --for=pull
  oc secrets link opssight-processor custom-registry-pull-secret --for=pull
  oc secrets link opssight-scanner custom-registry-pull-secret --for=pull
  oc scale rc opssight-core --replicas=0
  oc scale rc opssight-core --replicas=1
  oc scale rc opssight-scanner --replicas=0
  oc scale rc opssight-scanner --replicas=1
  oc scale rc opssight-pod-processor --replicas=0
  oc scale rc opssight-pod-processor --replicas=1
  oc scale rc opssight-image-processor --replicas=0
  oc scale rc opssight-image-processor --replicas=1
fi
