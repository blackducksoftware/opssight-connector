module github.com/blackducksoftware/opssight-connector

go 1.13

require (
	github.com/aws/aws-sdk-go v1.35.2
	github.com/blackducksoftware/hub-client-go v0.9.6 // indirect
	github.com/blackducksoftware/perceivers v2.2.4+incompatible
	github.com/blackducksoftware/perceptor v2.2.4+incompatible
	github.com/blackducksoftware/perceptor-scanner v2.2.3+incompatible
	github.com/blackducksoftware/perceptor-skyfire v2.2.3+incompatible
	github.com/blackducksoftware/synopsysctl v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty v0.0.0-20190619084753-e284be3e6edc // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	helm.sh/helm/v3 v3.1.1
	k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.3+incompatible
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.8.2
	github.com/blackducksoftware/hub-client-go => github.com/blackducksoftware/hub-client-go v0.9.5-0.20181018202023-f1fdde519aec
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.0-pre1
	github.com/ugorji/go => github.com/ugorji/go v1.1.7
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
)
