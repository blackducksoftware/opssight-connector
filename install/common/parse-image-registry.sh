#!/bin/bash

if [[ $_arg_developer_mode == "off" ]] ; then
  # Parse the JSON file on each platform version and assign the image name and tag name
  for i in {1..6}
  do
   name=$(cat "$1" | awk -F"[,:}]" '{for(i=1;i<=NF;i++){if($i~/\042'name'\042/){print $(i+1)}}}' | tr -d '"' | sed -n ${i}p | xargs)
   image=$(cat "$1" | awk -F"[,:}]" '{for(i=1;i<=NF;i++){if($i~/\042'image'\042/){print $(i+1)}}}' | tr -d '"' | sed -n ${i}p | xargs)
   tag=$(cat "$1" | awk -F"[,:}]" '{for(i=1;i<=NF;i++){if($i~/\042'tag'\042/){print $(i+1)}}}' | tr -d '"' | sed -n ${i}p | xargs)
   case "$name" in
     pod-perceiver)
      pod_perceiver_image="$image"
      pod_perceiver_tag="$tag"
      ;;
     image-perceiver)
      image_perceiver_image="$image"
      image_perceiver_tag="$tag"
      ;;
     perceptor)
      perceptor_image="$image"
      perceptor_tag="$tag"
      ;;
     perceptor-scanner)
      perceptor_scanner_image="$image"
      perceptor_scanner_tag="$tag"
      ;;
     perceptor-imagefacade)
      perceptor_imagefacade_image="$image"
      perceptor_imagefacade_tag="$tag"
      ;;
     perceptor-protoform)
      perceptor_protoform_image="$image"
      perceptor_protoform_tag="$tag"
      ;;
     *)
      Message="Ignore the pod:: $name."
      ;;
    esac
  done
fi
