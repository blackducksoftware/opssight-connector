#!/bin/bash

CPUS=$1

cat << EOF > opssight.yml
apiVersion: synopsys.com/v1
kind: OpsSight
metadata:
  clusterName: ""
  name: ops
spec:
  namespace: ops
  annotationIntervalSeconds: 30 # Optional
  dumpIntervalMinutes: 30 # Optional
  hubUser: sysadmin # Optional
  hubPort: 443 # Optional
  hubClientTimeoutPerceptorMilliseconds: 100000 # Optional
  hubClientTimeoutScannerSeconds: 600 # Optional
  concurrentScanLimit: 2 # Optional
  totalScanLimit: 1000 # Optional
  # CONTAINER PULL CONFIG
  perceptorImage: "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master"
  scannerImage: "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master"
  podPerceiverImage: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master"
  imageFacadeImage: "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master"
  skyfireImage: "gcr.io/saas-hub-stg/blackducksoftware/skyfire:master"
  imagePerceiver: false # OpenShift only!  Set to true to scan images in the OpenShift internal Docker registry
  podPerceiver: true  # Both Kubernetes and Openshift.  Set to true to scan images running in pods.
  metrics: true  # Set to true to enable a prometheus master for metrics visualizations.
  # Example: "300m"
  defaultCpu: 300m # Optional
  # Example: "1300Mi"
  defaultMem: 1300Mi # Optional
  # Log level
  logLevel: debug # Optional
  # Environment Variables
  hubuserPasswordEnvVar: PCP_HUBUSERPASSWORD # Optional, environment variable used to manage the Hub password
  # Configuration secret
  secretName: blackduck-secret # Optional, if the default perceptor secret name to be changed
  initialNoOfHubs: 0
  maxNoOfHubs: 0
  hubSpec:
    backupSupport: "No" # Required, possible values are 'Yes', 'No'
    certificateName: default # Required, possible values are 'default', 'manual' or other hub names
    dbPrototype: "empty" # Required, possible values are empty or other hub names
    dockerRegistry: docker.io # Required
    dockerRepo: blackducksoftware # Required
    hubVersion: 5.0.0 # Required
    flavor: small # Required, possible values are 'small', 'medium', 'large' or 'opssight'
    hubType: senthil # Required, possible values are 'master' or 'worker' or 'any custom value to filter the hubs corresponding to particular opssight'
EOF