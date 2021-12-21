module github.com/redhat-cop/vault-config-operator

go 1.16

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/go-logr/logr v0.4.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/hashicorp/vault/api v1.1.1
	github.com/masterminds/sprig v2.22.0+incompatible
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/redhat-cop/operator-utils v1.3.1
	github.com/scylladb/go-set v1.0.2
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
