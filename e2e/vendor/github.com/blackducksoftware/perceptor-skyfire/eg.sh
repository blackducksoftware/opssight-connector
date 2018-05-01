# create a dump from kubernetes, perceptor, and the hub
go run cmd/e2e/e2e.go cmd/e2e/config.json > dump4.txt

# process the dump into a report
python reports/report.py ./dump4.txt  > pyed2.txt

# cleanup kube: remove bd annotations and labels
go run cmd/cleanupkube/cleanupkube.go cmd/cleanupkube/config.json
