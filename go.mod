module code.cloudfoundry.org/quarks-utils

go 1.13

require (
	code.cloudfoundry.org/cf-operator v0.4.1
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.6.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.2
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
	sigs.k8s.io/yaml v1.1.0
)
