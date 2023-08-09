module github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.0
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/sys v0.0.0-20190429190828-d89cdac9e872 // indirect
	golang.org/x/text v0.3.2 // indirect
	istio.io/api v0.0.0-20190930220724-33a483a29b8e
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.15.7
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)
