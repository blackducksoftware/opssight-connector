#!/bin/bash
source ../common/parse-or-gather-user-input.sh "${@}"

_arg_image_perceiver="on"

oc new-project $_arg_pcp_namespace

oc project $_arg_pcp_namespace

source ../common/oadm-policy-init.sh $arg_pcp_namespace

source ../common/parse-image-registry.sh "../openshift/image-registry.json"

if [[ "$_arg_rhcc" == "on" ]] ; then
  oc create secret generic redhat-connect --from-file=.dockerconfigjson="$_arg_docker_config_path" --type=kubernetes.io/dockerconfigjson
  oc secrets link protoform redhat-connect --for=pull
  oc secrets link perceptor-scanner-sa redhat-connect --for=pull
  perceptor_image=$(echo "$perceptor_image-v2")
  perceptor_scanner_image=$(echo "$perceptor_scanner_image-v2")
  pod_perceiver_image=$(echo "$pod_perceiver_image-v2")
  image_perceiver_image=$(echo "$image_perceiver_image-v2")
  perceptor_imagefacade_image=$(echo "$perceptor_imagefacade_image-v2")
  perceptor_protoform_image=$(echo "$perceptor_protoform_image-v2")
fi

source ../common/protoform.yaml.sh

oc create -f protoform.yaml

if [[ $_arg_prometheus_metrics == "on" ]] ; then
  oc create -f ../common/prometheus-deployment.yaml
  oc expose service prometheus --port=9090 --name=prometheus-metrics
fi

rm -rf protoform.yaml
