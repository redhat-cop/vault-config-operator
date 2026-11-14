package olm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// etcdDefaultMaxRequestBytes is etcd's default --max-request-bytes.
// OLM packs the operator bundle into a single registry ConfigMap; if that
// object exceeds this limit, InstallPlan creation fails and DeployableByOLM times out.
const etcdDefaultMaxRequestBytes = 1572864

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to locate test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func TestOLMRegistryConfigMapUnderEtcdLimit(t *testing.T) {
	root := repoRoot(t)
	crdDir := filepath.Join(root, "config", "crd", "bases")
	csvPath := filepath.Join(root, "config", "manifests", "bases", "vault-config-operator.clusterserviceversion.yaml")
	samplesKustomization := filepath.Join(root, "config", "samples", "kustomization.yaml")
	bundleDir := filepath.Join(root, "bundle", "manifests")

	data := map[string]string{}

	crdEntries, err := os.ReadDir(crdDir)
	if err != nil {
		t.Fatalf("read CRD bases: %v", err)
	}
	for _, e := range crdEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(crdDir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		data[e.Name()] = string(content)
	}
	if len(data) == 0 {
		t.Fatal("no CRDs found in config/crd/bases")
	}

	almExamples, err := buildALMExamples(root, samplesKustomization)
	if err != nil {
		t.Fatalf("build alm-examples from samples: %v", err)
	}

	csvContent, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read CSV base: %v", err)
	}
	csvContent, err = injectALMExamples(csvContent, almExamples)
	if err != nil {
		t.Fatalf("inject alm-examples into CSV: %v", err)
	}
	data["vault-config-operator.clusterserviceversion.yaml"] = string(csvContent)

	// bundle/ is gitignored; when present locally, include the remaining small
	// non-CRD manifests OLM also packs into the registry ConfigMap.
	if bundleEntries, err := os.ReadDir(bundleDir); err == nil {
		for _, e := range bundleEntries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			if strings.HasPrefix(e.Name(), "redhatcop.redhat.io_") || strings.Contains(e.Name(), "clusterserviceversion") {
				continue
			}
			content, err := os.ReadFile(filepath.Join(bundleDir, e.Name()))
			if err != nil {
				t.Fatalf("read %s: %v", e.Name(), err)
			}
			data[e.Name()] = string(content)
		}
	}

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-config-operator-registry",
		},
		Data: data,
	}
	encoded, err := json.Marshal(cm)
	if err != nil {
		t.Fatalf("marshal ConfigMap: %v", err)
	}

	size := len(encoded)
	t.Logf("OLM registry ConfigMap JSON size: %d bytes (%.3f MiB); limit %d; objects=%d",
		size, float64(size)/(1024*1024), etcdDefaultMaxRequestBytes, len(data))

	if size >= etcdDefaultMaxRequestBytes {
		t.Fatalf("OLM registry ConfigMap is %d bytes, must stay under etcd limit of %d bytes (%.3f MiB). Trim CRD descriptions (crd:maxDescLen) or shrink alm-examples.",
			size, etcdDefaultMaxRequestBytes, float64(etcdDefaultMaxRequestBytes)/(1024*1024))
	}
}

func TestAllCRDsHaveBundleSamples(t *testing.T) {
	root := repoRoot(t)
	crdDir := filepath.Join(root, "config", "crd", "bases")
	samplesDir := filepath.Join(root, "config", "samples")
	kustomization := filepath.Join(samplesDir, "kustomization.yaml")

	listed, err := sampleResourcesFromKustomization(kustomization)
	if err != nil {
		t.Fatalf("parse samples kustomization: %v", err)
	}
	listedSet := map[string]struct{}{}
	for _, name := range listed {
		listedSet[name] = struct{}{}
	}

	sampleKinds := map[string]string{}
	for name := range listedSet {
		content, err := os.ReadFile(filepath.Join(samplesDir, name))
		if err != nil {
			t.Fatalf("read sample %s listed in kustomization: %v", name, err)
		}
		var obj struct {
			Kind string `json:"kind"`
		}
		if err := yaml.Unmarshal(content, &obj); err != nil {
			t.Fatalf("parse sample %s: %v", name, err)
		}
		if obj.Kind == "" {
			t.Fatalf("sample %s has empty kind", name)
		}
		sampleKinds[obj.Kind] = name
	}

	entries, err := os.ReadDir(crdDir)
	if err != nil {
		t.Fatalf("read CRD bases: %v", err)
	}
	var missing []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(crdDir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var crd struct {
			Spec struct {
				Names struct {
					Kind string `json:"kind"`
				} `json:"names"`
			} `json:"spec"`
		}
		if err := yaml.Unmarshal(content, &crd); err != nil {
			t.Fatalf("parse CRD %s: %v", e.Name(), err)
		}
		kind := crd.Spec.Names.Kind
		if kind == "" {
			t.Fatalf("CRD %s has empty kind", e.Name())
		}
		if _, ok := sampleKinds[kind]; !ok {
			missing = append(missing, kind)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("CRDs missing from config/samples kustomization (needed for alm-examples): %s", strings.Join(missing, ", "))
	}
}

func sampleResourcesFromKustomization(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var kus struct {
		Resources []string `json:"resources"`
	}
	if err := yaml.Unmarshal(content, &kus); err != nil {
		return nil, err
	}
	return kus.Resources, nil
}

func buildALMExamples(root, kustomization string) (string, error) {
	resources, err := sampleResourcesFromKustomization(kustomization)
	if err != nil {
		return "", err
	}
	samplesDir := filepath.Join(root, "config", "samples")
	examples := make([]json.RawMessage, 0, len(resources))
	for _, name := range resources {
		raw, err := os.ReadFile(filepath.Join(samplesDir, name))
		if err != nil {
			return "", err
		}
		var obj any
		if err := yaml.Unmarshal(raw, &obj); err != nil {
			return "", err
		}
		encoded, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		examples = append(examples, encoded)
	}
	out, err := json.Marshal(examples)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func injectALMExamples(csvYAML []byte, almExamples string) ([]byte, error) {
	var obj map[string]any
	if err := yaml.Unmarshal(csvYAML, &obj); err != nil {
		return nil, err
	}
	metadata, _ := obj["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
		obj["metadata"] = metadata
	}
	annotations, _ := metadata["annotations"].(map[string]any)
	if annotations == nil {
		annotations = map[string]any{}
		metadata["annotations"] = annotations
	}
	annotations["alm-examples"] = almExamples
	return yaml.Marshal(obj)
}
