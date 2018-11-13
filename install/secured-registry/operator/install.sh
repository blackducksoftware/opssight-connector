#!/bin/bash

unset DYLD_INSERT_LIBRARIES

source `dirname ${BASH_SOURCE}`/args.sh "${@}"

if [[ -z "$_arg_registration_key" ]]; then
  echo "please provide the Black Duck registration key!!!"
  exit 1
fi

echo "Using the secret encoded in this file.  Change it before running, or press enter..."
read x

cat << EOF > /tmp/secret
api_arg_version: v1
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

oc new-project "$_arg_namespace"

oc create -f /tmp/secret -n "$_arg_namespace"

cat ../../blackduck-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$_arg_registration_key'/g' | \
sed 's/${NAMESPACE}/'$_arg_namespace'/g' | \
sed 's/${TAG}/'$_arg_version'/g' | \
sed 's/${DOCKER_REGISTRY}/'$_arg_registry'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $_arg_project | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace="$_arg_namespace" -f -

if [[ ! -z "$_arg_docker_config" ]]; then
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$_arg_docker_config" --type=kubernetes.io/dockerconfigjson
  oc secrets link default custom-registry-pull-secret --for=pull
  oc secrets link blackduck-operator custom-registry-pull-secret --for=pull; 
  oc scale rc blackduck-operator --replicas=0
  oc scale rc blackduck-operator --replicas=1
fi

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-np --type=NodePort --namespace="$_arg_namespace"

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-lb --type=LoadBalancer --namespace="$_arg_namespace"
