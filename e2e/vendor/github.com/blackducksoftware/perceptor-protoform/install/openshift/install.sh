#!/bin/bash
source ../common/parse-or-gather-user-input.sh "${@}"

_arg_image_perceiver="on"

oc new-project $_arg_pcp_namespace

source ../common/oadm-policy-init.sh $arg_pcp_namespace

source ../common/parse-image-registry.sh "../openshift/image-registry.json"

source ../common/protoform.yaml.sh
#oc project $_arg_pcp_namespace
oc create -f protoform.yaml

if [[ $_arg_prometheus_metrics == "on" ]] ; then
  oc create -f ../common/prometheus-deployment.yaml
  oc expose service prometheus --port=9090 --name=prometheus-metrics
fi
