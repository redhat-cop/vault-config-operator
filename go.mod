module github.com/redhat-cop/vault-config-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/hashicorp/vault/api v1.1.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.1.2 // indirect
	github.com/onsi/gomega v1.18.1
	github.com/redhat-cop/operator-utils v1.3.1
	github.com/scylladb/go-set v1.0.2
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
