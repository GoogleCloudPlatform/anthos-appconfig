module github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/net v0.0.0-20190501004415-9ce7a6920f09
	istio.io/api v0.0.0-20190620164021-251533346f33
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
	sigs.k8s.io/controller-tools v0.2.0-beta.2 // indirect
)
