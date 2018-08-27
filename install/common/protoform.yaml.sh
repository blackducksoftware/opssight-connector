#!/bin/bash

openshift="false"
if [[ "$_arg_image_perceiver" == "on" ]] ; then
  openshift="true"
fi

skyfire="false"
if [[ "$_arg_skyfire" == "on" ]] ; then
  skyfire="true"
fi

DEF_PERCEPTOR_PROTOFORM_IMAGE=perceptor-protoform
DEF_PERCEPTOR_PROTOFORM_TAG=master

perceptor_protoform_image=${perceptor_protoform_image:-$DEF_PERCEPTOR_PROTOFORM_IMAGE}
perceptor_protoform_tag=${perceptor_protoform_tag:-$DEF_PERCEPTOR_PROTOFORM_TAG}

perceptor_protoform_tag=${_arg_default_container_version:-$perceptor_protoform_tag}
perceptor_tag=${_arg_default_container_version:-$perceptor_tag}
perceptor_scanner_tag=${_arg_default_container_version:-$perceptor_scanner_tag}
pod_perceiver_tag=${_arg_default_container_version:-$pod_perceiver_tag}
perceptor_imagefacade_tag=${_arg_default_container_version:-$perceptor_imagefacade_tag}

hubUserPassword=$(printf "%s" "$_arg_hub_password" | base64)

cat << EOF > protoform.yaml
apiVersion: v1
kind: Pod
metadata:
  name: protoform
spec:
  volumes:
  - name: protoform
    configMap:
      name: protoform
  containers:
  - name: protoform
    image: ${_arg_container_registry}/${_arg_image_repository}/${perceptor_protoform_image}:${perceptor_protoform_tag}
    env:
    - name: PCP_HUBUSERPASSWORD
      valueFrom:
        secretKeyRef:
          name: viper-secret
          key: HubUserPassword
    imagePullPolicy: Always
    command: [ ./protoform ]
    args: ["/etc/protoform/protoform.yaml"]
    ports:
    - containerPort: 3001
      protocol: TCP
    volumeMounts:
    - name: protoform
      mountPath: /etc/protoform/
  restartPolicy: Never
  serviceAccountName: protoform
  serviceAccount: protoform
---
apiVersion: v1
kind: List
metadata:
  name: viper-inputs
items:
- apiVersion: v1
  kind: Secret
  metadata:
    name: viper-secret
  type: Opaque
  data:
    HubUserPassword: "$hubUserPassword"
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: protoform
  data:
    protoform.yaml: |
      HubHost: "$_arg_hub_host"
      HubUser: "$_arg_hub_user"
      HubPort: "$_arg_hub_port"
      HubClientTimeoutPerceptorSeconds: "$_arg_hub_client_timeout_perceptor_seconds"
      HubClientTimeoutScannerSeconds: "$_arg_hub_client_timeout_scanner_seconds"
      ConcurrentScanLimit: "$_arg_hub_max_concurrent_scans"
      Namespace: "$_arg_pcp_namespace"
      Openshift: "$openshift"
      InternalRegistries: '`echo "$_arg_private_registry"`'
      DefaultCPU: "$_arg_container_default_cpu"
      DefaultMem: "$_arg_container_default_memory"
      AnnotationIntervalSeconds: "$_arg_annotation_interval_seconds"

      # TODO: Assuming for now that we run the same version of everything
      # For the curated installers.  For developers ? You might want to
      # hard code one of these values if using this script for dev/test.
      Registry: "$_arg_container_registry"
      ImagePath: "$_arg_image_repository"
      Defaultversion: "$_arg_default_container_version"
      PerceptorImageName: "$perceptor_image"
      ScannerImageName: "$perceptor_scanner_image"
      PodPerceiverImageName: "$pod_perceiver_image"
      ImagePerceiverImageName: "$image_perceiver_image"
      ImageFacadeImageName: "$perceptor_imagefacade_image"
      PerceptorContainerVersion: "$perceptor_tag"
      ScannerContainerVersion: "$perceptor_scanner_tag"
      PerceiverContainerVersion: "$pod_perceiver_tag"
      ImageFacadeContainerVersion: "$perceptor_imagefacade_tag"
      LogLevel: "$_arg_container_default_log_level"
      PerceptorSkyfire: "$skyfire"
EOF
