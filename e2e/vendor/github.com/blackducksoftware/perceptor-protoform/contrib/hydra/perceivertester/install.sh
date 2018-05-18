#!/bin/bash

set +x
NS=bds-perceptor

POD_PERCEIVER_SA="pod-perceiver-sa"
IMAGE_PERCEIVER_SA="image-perceiver-sa"

cleanup() {
	oc delete ns $NS
	while oc get ns | grep -q $NS ; do
	  echo "Waiting for deletion...`oc get ns | grep $NS` "
	  sleep 1
	done
}

install-rbac() {
		set -e

		oc new-project $NS

		oc create serviceaccount $POD_PERCEIVER_SA -n $NS
		oc create serviceaccount $IMAGE_PERCEIVER_SA -n $NS
		# allows writing of cluster level metadata for imagestreams
		oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:$NS:$POD_PERCEIVER_SA
		oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:$NS:$IMAGE_PERCEIVER_SA
}

cleanup
install-rbac


cat << EOF > aux-config.json
{
	"Namespace": "$NS",
	"PodPerceiverServiceAccountName": "$POD_PERCEIVER_SA",
	"ImagePerceiverServiceAccountName": "$IMAGE_PERCEIVER_SA",
	"IsOpenshift": true
}
EOF

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./perceivertester.go ./config.json aux-config.json
