#!/bin/bash

export SCC="add-scc-to-user"
export ROLE="add-role-to-user"
export CLUSTER="add-cluster-role-to-user"

source protoform.yaml.sh
source rbac.yaml.sh

kubectl create ns $NS
kubectl create -f  /tmp/protoform-rbac.yaml -n $NS
kubectl create sa perceptor-scanner -n $NS
kubectl create sa perceiver -n $NS
