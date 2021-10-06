module github.com/redhat-cop/vault-config-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/hashicorp/vault/api v1.1.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/scylladb/go-set v1.0.2
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)
