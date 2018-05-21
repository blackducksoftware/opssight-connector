set -e

# assume perceptor has already been spun up

sleepseconds=1800

# wait 30 minutes
echo "about to sleep $sleepseconds seconds .... "
sleep $sleepseconds

# verify skyfire stats
echo "about to run basic skyfire tests"
set +e
ginkgo ./basicskyfire -- --skyfireHost $1 --skyfirePort $2
set -e

# spin up new pod
set +e
echo "create namespace"
kubectl create ns random
set -e
echo "running new pod"
kubectl run echoer-2 --image gcr.io/gke-verification/blackducksoftware/echoer:2 -n random

# wait 30 minutes
echo "about to sleep $sleepseconds seconds .... "
sleep $sleepseconds

# verify skyfire tests
echo "about to run basic skyfire tests"
set +e
ginkgo ./basicskyfire -- --skyfireHost $1 --skyfirePort $2
set -e

# ???? tests
# go test ./pkg

# clean up
echo "deleting namespace"
kubectl delete ns random
