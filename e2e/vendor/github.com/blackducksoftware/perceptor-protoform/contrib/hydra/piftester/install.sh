#!/bin/bash

set +x
NS=bds-perceptor
KUBECTL="kubectl"

IMAGEFACADE_SA="imagefacade-sa"
IMAGE_PERCEIVER_SA="image-perceiver-sa"
POD_PERCEIVER_SA="pod-perceiver-sa"

function is_openshift {
	if `which oc` ; then
		# oc version
		return 0
	else
		return 1
	fi
	return 1
}

cleanup() {
	is_openshift
	if ! $(exit $?); then
		echo "assuming kube"
		KUBECTL="kubectl"
	else
		KUBECTL="oc"
	fi
	$KUBECTL delete ns $NS
	while $KUBECTL get ns | grep -q $NS ; do
	  echo "Waiting for deletion...`$KUBECTL get ns | grep $NS` "
	  sleep 1
	done
}

install-rbac() {
	if [ "$KUBECTL" == "kubectl" ]; then
		echo "Detected Kubernetes... setting up"
		kubectl create ns $NS
		kubectl create sa $IMAGEFACADE_SA -n $NS
		kubectl create sa $POD_PERCEIVER_SA -n $NS
  else
		set -e

		echo "Detected openshift... setting up "
		oc new-project $NS

		oc create serviceaccount $IMAGEFACADE_SA -n $NS
		# allows launching of privileged containers for Docker machine access
		oc adm policy add-scc-to-user privileged system:serviceaccount:$NS:$IMAGEFACADE_SA
		# Allows pulling, viewing all images
		oc policy add-role-to-user view system:serviceaccount:$NS:$IMAGEFACADE_SA

		# allows pulling of images
		oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:$NS:$IMAGEFACADE_SA
		# oc adm policy add-role-to-user system:image-puller sytem:serviceaccount:$NS:$IMAGEFACADE_SA
		# oc adm policy add-role-to-user admin system:serviceaccount:$NS:$IMAGEFACADE_SA -n openshift
		# oc adm policy add-role-to-user system:registry sytem:serviceaccount:$NS:$IMAGEFACADE_SA

		oc create serviceaccount $POD_PERCEIVER_SA -n $NS
		oc create serviceaccount $IMAGE_PERCEIVER_SA -n $NS
		# allows writing of cluster level metadata for imagestreams
		oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:$NS:$POD_PERCEIVER_SA
		oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:$NS:$IMAGE_PERCEIVER_SA
	fi
}

cleanup
install-rbac


## finished initial setup, now run piftester

DOCKER_PASSWORD=$(oc sa get-token $IMAGEFACADE_SA)

## TODO: stop hardcoding internal docker registries

cat << EOF > aux-config.json
{
	"Namespace": "$NS",
	"DockerUsername": "admin",
	"DockerPassword": "$DOCKER_PASSWORD",
	"PodPerceiverServiceAccountName": "$POD_PERCEIVER_SA",
	"ImagePerceiverServiceAccountName": "$IMAGE_PERCEIVER_SA",
	"ImageFacadeServiceAccountName": "$IMAGEFACADE_SA",
	"InternalDockerRegistries": [
		"docker-registry.default.svc:5000",
		"172.30.28.16:5000"
	]
}
EOF

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./piftester.go ./config.json aux-config.json
