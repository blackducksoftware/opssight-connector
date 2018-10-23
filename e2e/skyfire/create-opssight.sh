#!/bin/bash

CPUS=$1

cat << EOF > create-opssight.yml
apiVersion: synopsys.com/v1
kind: OpsSight
metadata:
  clusterName: ""
  name: opssight-test
  namespace: ""
spec:
  annotationIntervalSeconds: 30
  checkForStalledScansPauseHours: 999999
  concurrentScanLimit: 2
  defaultCpu: ${CPUS}
  defaultMem: 1300Mi
  defaultVersion: master
  dumpIntervalMinutes: 30
  hubClientTimeoutPerceptorMilliseconds: 100000
  hubClientTimeoutScannerSeconds: 600
  hubPort: 443
  hubUser: sysadmin
  hubuserPasswordEnvVar: PCP_HUBUSERPASSWORD
  imageFacadeImageName: perceptor-imagefacade
  imageFacadePort: 3004
  imagePath: saas-hub-stg/blackducksoftware
  imagePerceiver: false
  imagePerceiverImageName: image-perceiver
  logLevel: info
  metrics: true
  modelMetricsPauseSeconds: 15
  names:
    image-perceiver: image-perceiver
    perceiver: perceiver
    perceptor: perceptor
    perceptor-image-facade: perceptor-imagefacade
    perceptor-scanner: perceptor-scanner
    pod-perceiver: pod-perceiver
    skyfire: skyfire
  namespace: opssight-test
  perceiverPort: 3002
  perceptorImageName: perceptor
  perceptorPort: 3001
  perceptorSkyfire: false
  podPerceiver: true
  podPerceiverImageName: pod-perceiver
  registry: gcr.io
  scannerImageName: perceptor-scanner
  scannerPort: 3003
  secretName: blackduck-secret
  serviceAccounts:
    image-perceiver: perceiver
    perceptor-image-facade: perceptor-scanner
    pod-perceiver: perceiver
    skyfire: skyfire
  skyfireImageName: skyfire
  skyfirePort: 3005
  stalledScanClientTimeoutHours: 999999
  state: running
  totalScanLimit: 1000
  unknownImagePauseMilliseconds: 15000
  useMockMode: false
EOF