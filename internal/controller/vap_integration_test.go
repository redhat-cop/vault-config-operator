//go:build integration

/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// vapManifestObjects holds the VAP resources deployed for this test suite
// so they can be cleaned up afterwards.
var vapManifestObjects []*unstructured.Unstructured

var _ = Describe("VAP Integration", Ordered, func() {

	defaultAuth := func() vaultutils.KubeAuthConfiguration {
		return vaultutils.KubeAuthConfiguration{
			Path:           "kubernetes",
			Role:           "policy-admin",
			ServiceAccount: &corev1.LocalObjectReference{Name: "default"},
		}
	}

	BeforeAll(func() {
		By("deploying VAP manifests from config/vap/")
		vapDir := filepath.Join("..", "..", "config", "vap")
		vapFiles, err := filepath.Glob(filepath.Join(vapDir, "*.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(vapFiles).NotTo(BeEmpty(), "no VAP manifests found in config/vap/")

		for _, vapFile := range vapFiles {
			if filepath.Base(vapFile) == "kustomization.yaml" {
				continue
			}
			data, err := os.ReadFile(vapFile)
			Expect(err).NotTo(HaveOccurred())
			reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
			for {
				doc, err := reader.Read()
				if err == io.EOF {
					break
				}
				Expect(err).NotTo(HaveOccurred())
				if len(bytes.TrimSpace(doc)) == 0 {
					continue
				}
				obj := &unstructured.Unstructured{}
				Expect(yaml.NewYAMLOrJSONDecoder(bytes.NewReader(doc), 4096).Decode(obj)).To(Succeed())
				err = k8sIntegrationClient.Create(ctx, obj)
				if apierrors.IsAlreadyExists(err) {
					// Track pre-existing objects so AfterAll cleans them up
					vapManifestObjects = append(vapManifestObjects, obj)
					continue
				}
				Expect(err).NotTo(HaveOccurred())
				vapManifestObjects = append(vapManifestObjects, obj)
			}
		}

		By("waiting for VAP compilation and enforcement")
		time.Sleep(5 * time.Second)
	})

	AfterAll(func() {
		By("cleaning up VAP manifests")
		for _, obj := range vapManifestObjects {
			_ = k8sIntegrationClient.Delete(ctx, obj)
		}
		vapManifestObjects = nil
	})

	Context("VAP manifests deployed", func() {
		It("should have ValidatingAdmissionPolicy resources in the cluster", func() {
			vapList := &admissionregistrationv1.ValidatingAdmissionPolicyList{}
			Expect(k8sIntegrationClient.List(ctx, vapList)).To(Succeed())

			vapNames := make([]string, 0, len(vapList.Items))
			for _, vap := range vapList.Items {
				vapNames = append(vapNames, vap.Name)
			}
			_, _ = fmt.Fprintf(GinkgoWriter, "Found %d VAPs in cluster: %v\n", len(vapNames), vapNames)

			Expect(vapNames).To(ContainElement("vco-immutable-path"))
			Expect(vapNames).To(ContainElement("vco-immutable-name"))
			Expect(vapNames).To(ContainElement("vco-target-namespace-guard"))
		})
	})

	Context("immutable-path enforcement in Kind", func() {
		It("should reject spec.path change for AuthEngineMount", func() {
			obj := &redhatcopv1alpha1.AuthEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "vap-int-path-",
					Namespace:    vaultAdminNamespaceName,
				},
				Spec: redhatcopv1alpha1.AuthEngineMountSpec{
					Authentication: defaultAuth(),
					Path:           "auth/vap-int-test",
					AuthMount:      redhatcopv1alpha1.AuthMount{Type: "approle", Description: "integration test"},
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, obj)).To(Succeed())
			DeferCleanup(func() { _ = k8sIntegrationClient.Delete(ctx, obj) })

			Eventually(func(g Gomega) {
				g.Expect(k8sIntegrationClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())
				obj.Spec.Path = "auth/vap-int-changed"
				err := k8sIntegrationClient.Update(ctx, obj)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("spec.path cannot be updated"))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("CRD-specific VAP enforcement in Kind", func() {
		It("should reject TransitSecretEngineKey with convergentEncryption but no derived", func() {
			obj := &redhatcopv1alpha1.TransitSecretEngineKey{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "vap-int-transit-",
					Namespace:    vaultAdminNamespaceName,
				},
				Spec: redhatcopv1alpha1.TransitSecretEngineKeySpec{
					Authentication: defaultAuth(),
					Path:           "transit",
					TransitKeyConfig: redhatcopv1alpha1.TransitKeyConfig{
						ConvergentEncryption: true,
						Derived:              false,
					},
				},
			}
			err := k8sIntegrationClient.Create(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("convergentEncryption requires"))
		})
	})
})
