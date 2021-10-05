module github.com/redhat-cop/vault-config-operator

go 1.16

require (
	github.com/hashicorp/vault/api v1.1.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)
