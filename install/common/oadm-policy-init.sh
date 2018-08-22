#!/bin/bash
SCC="add-scc-to-user"
ROLE="add-role-to-user"
CLUSTER="add-cluster-role-to-user"
NS=$_arg_pcp_namespace

# Create the perceiver service account
oc create serviceaccount perceiver -n $NS

# Protoform has its own SA.
oc create serviceaccount protoform -n $NS
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:protoform

# following allows us to write cluster level metadata for imagestreams
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceiver

# Create the serviceaccount for perceptor-scanner to talk with Docker
oc create sa perceptor-scanner -n $NS

# allows launching of privileged containers for Docker machine access
oc adm policy $SCC privileged system:serviceaccount:$NS:perceptor-scanner

# these allow us to pull images
oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceptor-scanner

oc policy $ROLE view system:serviceaccount::perceptor-scanner

_arg_private_registry_token=$(oc sa get-token perceptor-scanner)

# Get the default Docker Registry
route_docker_registry=$(oc get route docker-registry -n default -o jsonpath='{.spec.host}')
service_docker_registry=$(oc get svc docker-registry -n default -o jsonpath='{.spec.clusterIP}')
service_docker_registry_port=$(oc get svc docker-registry -n default -o jsonpath='{.spec.ports[0].port}')
_arg_private_registry="[{\"Url\": \"docker-registry.default.svc:5000\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}, {\"Url\": \"$route_docker_registry\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}, {\"Url\": \"$route_docker_registry:443\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}, {\"Url\": \"$service_docker_registry:$service_docker_registry_port\", \"User\": \"admin\", \"Password\": \"$_arg_private_registry_token\"}]"

# Note : This privileged SCC allows the perceivers to acto on accounts which have privileged ACLs.
# This is necessary, for example, for annotating a privileged pod!
oc adm policy add-scc-to-user privileged system:serviceaccount:$NS:perceiver
