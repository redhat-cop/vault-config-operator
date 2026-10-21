# Contributing a new Vault API

All vault APIs can be manipulated using the Vault logical client and essentially with three operations: `read`, `write` (corresponding to create and update) and `delete`.

A framework has been created to make it simple to add new types (essentially new secret engine configuration and role types and new authentication engine configuration and role types).

## Prerequisites

Before you begin, make sure you have:

- **Go** — the version declared in `go.mod` (currently Go 1.26+). Using a different Go version may cause `controller-gen` or other tools to fail.
- **operator-sdk** — the project pins a specific version in the Makefile (currently `v1.42.3`). Run `make operator-sdk` to download it to `./bin/operator-sdk`. **Do not use a globally-installed `operator-sdk`** — version mismatches with the project layout (`go.kubebuilder.io/v4`) will cause scaffolding failures.
- **controller-gen** — likewise pinned in the Makefile (currently `v0.21.0`). Run `make controller-gen` to download it to `./bin/controller-gen`. The scaffolding step and `make generate` / `make manifests` use this automatically.

> **Tip:** All tool versions are declared near the top of the `Makefile` (`OPERATOR_SDK_VERSION`, `CONTROLLER_TOOLS_VERSION`, etc.). The Makefile downloads each tool to `./bin/` on first use, so you generally don't need to install anything globally.

## Steps

1. Download the pinned `operator-sdk` and scaffold the CRD type, controller, and webhooks:

   ```shell
   make operator-sdk
   ./bin/operator-sdk create api --group redhatcop --version v1alpha1 --kind MyVaultType --resource --controller
   ./bin/operator-sdk create webhook --group redhatcop --version v1alpha1 --kind MyVaultType --defaulting --programmatic-validation
   ```

2. Define the Vault type fields with `vault:` struct tags. You should define the vault-specific fields in an inline type. Each exported field that maps to a Vault API parameter gets a `vault:"<key>"` tag. Use `vault:"-"` for fields to exclude (Connection, Authentication, Path, credential structs). Here is an example:

   ```go
   type MyVaultTypeSpec struct {
       // Connection represents the information needed to connect to Vault.
       // +kubebuilder:validation:Optional
       Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

       // Authentication is the kube auth configuration to be used to execute this request
       // +kubebuilder:validation:Required
       Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

       // Path at which to create the role.
       // +kubebuilder:validation:Required
       Path vaultutils.Path `json:"path,omitempty"`

       MyVaultSpecificFields `json:",inline"`

       // The name of the object created in Vault.
       // +kubebuilder:validation:Optional
       Name string `json:"name,omitempty"`
   }

   type MyVaultSpecificFields struct {
       NormalField string `vault:"normal_field" json:"normalField,omitempty"`
       TTLField    string `vault:"ttl,duration" json:"ttl,omitempty"`
       Count       int    `vault:"count,number" json:"count,omitempty"`
   }
   ```

   ### Vault Struct Tag Grammar

   ```
   vault:"<vault_key>[,<codec>][,omitempty][,when=Field:value][,default=X][,spreadmap]"
   vault:"-"          // Exclude this field entirely
   vault:",inline"    // Merge embedded struct fields into parent
   ```

   **Codecs (closed set):**

   | Codec | Input Types | Output | Example |
   |---|---|---|---|
   | _(identity/default)_ | `string`, `bool`, `[]string` | pass-through; `[]string` → `[]any` | `vault:"allowed_roles"` |
   | `duration` | `string` (Go duration) | `json.Number` (seconds) | `vault:"token_ttl,duration"` |
   | `number` | `int`, `int64` | `json.Number` | `vault:"token_num_uses,number"` |
   | `intmap` | `map[string]int` | `map[string]any{json.Number}` | `vault:"allowed_user_key_lengths,intmap"` |
   | `stringmap` | `map[string]string` | `map[string]any` | `vault:"default_extensions,stringmap"` |

   **Directives:**

   | Directive | Description | Use Case |
   |---|---|---|
   | `omitempty` | Skip when zero value | Optional fields |
   | `when=Field:value` | Include only when sibling Go field == value | SSH `key_type`, TOTP `generate` |
   | `default=X` | Emit X when pointer is nil | `*bool verify_connection`, `*int max_retries` |
   | `spreadmap` | Merge map entries into parent payload | DB `DatabaseSpecificConfig` |
   | `inline` | Recurse into embedded struct | All inline config structs |
   | `-` | Skip field entirely | Connection, Authentication, Path |

3. Implement the `VaultObject` interface (5 required methods):

   ```go
   var _ vaultutils.VaultObject = &MyVaultType{}

   func (d *MyVaultType) GetPath() string {
       return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
   }

   func (d *MyVaultType) GetPayload() map[string]any {
       return vaultutils.StructToVaultPayload(&d.Spec.MyVaultSpecificFields)
   }

   func (d *MyVaultType) IsDeletable() bool { return true }

   func (d *MyVaultType) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
       return &d.Spec.Authentication
   }

   func (d *MyVaultType) GetVaultConnection() *vaultutils.VaultConnection {
       return d.Spec.Connection
   }
   ```

   **Do NOT hand-write `toMap()` methods** for new types. Use `vault:` struct tags + `StructToVaultPayload` instead. The only exception is `TransitSecretEngineKey` which produces two different maps from one struct.

   ### Optional Interfaces

   Only implement these if your type needs non-trivial logic. When absent, sensible defaults are used:

   | Interface | Default | Implement When |
   |---|---|---|
   | `VaultEquivalenceChecker` | `filterPayloadToDesiredKeys` + `reflect.DeepEqual` | Custom drift detection (secret-stripping, field remapping, etc.) |
   | `VaultValidator` | `(true, nil)` | Type-level validation (e.g., mutually exclusive fields) |
   | `VaultInternalValuesPreparer` | no-op | Credential resolution, accessor lookup, password generation |
   | `VaultTLSPreparer` | no-op | Not currently used by any type |

4. Implement the `vaultutils.ConditionsAware` interface:

   ```go
   var _ vaultutils.ConditionsAware = &MyVaultType{}
   ```

5. Add needed validation and defaulting to the webhook. Use the generic `admission.Defaulter[*T]` / `admission.Validator[*T]` signatures. All resources should prevent the path from being changed:

   ```go
   var _ admission.Defaulter[*MyVaultType] = &MyVaultType{}

   func (r *MyVaultType) Default(ctx context.Context, obj *MyVaultType) error {
       return nil
   }

   var _ admission.Validator[*MyVaultType] = &MyVaultType{}

   func (r *MyVaultType) ValidateCreate(ctx context.Context, obj *MyVaultType) (admission.Warnings, error) {
       return nil, nil
   }

   func (r *MyVaultType) ValidateUpdate(ctx context.Context, oldObj, newObj *MyVaultType) (admission.Warnings, error) {
       if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
           return nil, err
       }
       if newObj.Spec.Path != oldObj.Spec.Path {
           return nil, errors.New("spec.path cannot be updated")
       }
       return nil, nil
   }

   func (r *MyVaultType) ValidateDelete(ctx context.Context, obj *MyVaultType) (admission.Warnings, error) {
       return nil, nil
   }
   ```

6. Implement the controller — under normal circumstances, just use this pattern:

   ```go
   type MyVaultTypeReconciler struct {
       vaultresourcecontroller.ReconcilerBase
   }

   func (r *MyVaultTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       instance := &redhatcopv1alpha1.MyVaultType{}
       err := r.GetClient().Get(ctx, req.NamespacedName, instance)
       if err != nil {
           if apierrors.IsNotFound(err) {
               return reconcile.Result{}, nil
           }
           return reconcile.Result{}, err
       }

       ctx1, err := prepareContext(ctx, r.ReconcilerBase, instance)
       if err != nil {
           r.Log.Error(err, "unable to prepare context", "instance", instance)
           return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
       }
       vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)
       return vaultResource.Reconcile(ctx1, instance)
   }
   ```

7. On the `main.go` update the controller reconciler and webhooks.

8. Regenerate the CRD manifests, RBAC, and DeepCopy methods:

   ```shell
   make generate manifests
   ```

9. Build and test:

   ```shell
   make build test
   ```
