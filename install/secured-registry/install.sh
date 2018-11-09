#!/bin/bash
echo "args = Namespace, Reg_key, version, docker config json path, docker registry(not mandatory) and docker project(not mandatory)"

NS=$1
REG_KEY=$2
VERSION=$3
DOCKER_CONFIG_JSON_PATH="$4"

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

if [[ ! -z "$5" ]]; then
  DOCKER_REGISTRY="$5"
else
  DOCKER_REGISTRY="registry.connect.redhat.com"
fi 

if [[ ! -z "$6" ]]; then
  DOCKER_REPO="$6"
else
  DOCKER_REPO="blackducksoftware"
fi

cat ../blackduck-operator.yaml | \
sed 's/${REGISTRATION_KEY}/'$REG_KEY'/g' | \
sed 's/${NAMESPACE}/'$NS'/g' | \
sed 's/${TAG}/'${VERSION}'/g' | \
sed 's/${DOCKER_REGISTRY}/'$DOCKER_REGISTRY'/g' | \
sed 's/${DOCKER_REPO}/'$(echo $DOCKER_REPO | sed -e 's/\\/\\\\/g; s/\//\\\//g; s/&/\\\&/g')'/g' | \
oc create --namespace=$NS -f -

if [[ ! -z "$DOCKER_CONFIG_JSON_PATH" ]]; then
  oc create secret generic custom-registry-pull-secret --from-file=.dockerconfigjson="$DOCKER_CONFIG_JSON_PATH" --type=kubernetes.io/dockerconfigjson
  oc secrets link default custom-registry-pull-secret --for=pull
  oc secrets link blackduck-operator custom-registry-pull-secret --for=pull; 
  oc scale rc blackduck-operator --replicas=0
  oc scale rc blackduck-operator --replicas=1
fi

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-np --type=NodePort --namespace=$NS

#oc expose rc blackduck-protoform --port=8080 --target-port=8080 --name=blackduck-protoform-lb --type=LoadBalancer --namespace=$NS
