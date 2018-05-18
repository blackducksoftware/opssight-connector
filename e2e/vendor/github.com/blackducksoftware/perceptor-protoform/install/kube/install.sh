#!/bin/bash

source ../common/parse-or-gather-user-input.sh "${@}"

kubectl create ns $_arg_pcp_namespace

source ../common/rbac.yaml.sh

source ../common/parse-image-registry.sh "../kube/image-registry.json"

source ../common/protoform.yaml.sh

set -x
kubectl create -f rbac.yaml -n $_arg_pcp_namespace
kubectl create -f protoform.yaml -n $_arg_pcp_namespace

if [[ "$_arg_prometheus_metrics" == "on" ]] ; then
  kubectl create -f ../common/prometheus-deployment.yaml -n $_arg_pcp_namespace
  kubectl expose service prometheus --port=9090 --name=prometheus-metrics -n $_arg_pcp_namespace
fi
