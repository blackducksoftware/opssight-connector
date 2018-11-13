#!/bin/bash

unset DYLD_INSERT_LIBRARIES

source `dirname ${BASH_SOURCE}`/args.sh "${@}"

cat ./opssight_with_min_parameter.yaml | \
sed 's/${NAMESPACE}/'$_arg_namespace'/g' | \
sed 's/${TAG}/'$_arg_version'/g' | \
sed 's/${DOCKER_REGISTRY}/'$_arg_registry'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $_arg_project | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create -f -

sleep 10;

if [[ ! -z "$_arg_docker_config" ]]; then
  oc project "$_arg_namespace"
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$_arg_docker_config" --type=kubernetes.io/dockerconfigjson
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
