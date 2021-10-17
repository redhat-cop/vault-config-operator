/*
Copyright 2021.

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

package vaultresourcecontroller

import (
	"strings"

	"github.com/redhat-cop/operator-utils/pkg/util"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getFinalizer(instance client.Object) string {
	return "controller-" + strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
}

func isValid(obj client.Object) (bool, error) {
	return obj.(vaultutils.VaultObject).IsValid()
}

func isInitialized(obj client.Object) bool {
	isInitialized := true
	if !util.HasFinalizer(obj, getFinalizer(obj)) {
		util.AddFinalizer(obj, getFinalizer(obj))
		isInitialized = false
	}
	return isInitialized || obj.(vaultutils.VaultObject).IsInitialized()
}
