go test -c -o ./basictest ./basicskyfire/

./basictest --skyfireHost skyfire --skyfirePort 3005 --noOfPods 40 --configPath /etc/mirage/mirage-config.yaml
