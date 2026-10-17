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

2. Define the My Vault type fields. You should define the vault specific field in an inline type. You should have a one to one mapping between the CRD fields and the Vault api field, except for field with special treatment. Add the path and authentication fields. Add any special treatment field as un-exported and un-marshalled fields. Here is an example:
  
  ```golang
    type MyVaultTypeSpec struct {

    // Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`      

    // Authentication is the kube aoth configuraiton to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to create the role.
    // The final path will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
    // The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
    // +kubebuilder:validation:Required
    Path Path `json:"path,omitempty"`

    MyVaultSpecificFields `json:",inline"`
    }

    type MyVaultSpecificFields struct {
      NormalField string `json:"normalField,omitempty"`
      specialField string `json:"-"`
    }
  ```

3. Implement the `VaultObejct` interface  

    ```golang
    var _ vaultutils.VaultObject = &MyVaultType{}
    ```

  then add the required methods. If you have special treatment field use the `PrepareInternalValues` method to initialze those fields.

4. Implements the `vaultutils.ConditionsAware` interface
  
   ```golang
   var _ vaultutils.ConditionsAware = &MyVaultType{}
   ```

5. Add needed validation and defaulting to the webhook. Notice that all the resources will need to prevent the path from being changed in the valitating webhooks:

  ```golang:
  func (r *MyVaultType) Default() {
    authenginemountlog.Info("default", "name", r.Name)
    //add your defaults here
  }

  func (r *MyVaultType) ValidateUpdate(old runtime.Object) error {
    authenginemountlog.Info("validate update", "name", r.Name)

    // the path cannot be updated
    if r.Spec.Path != old.(*MyVaultType).Spec.Path {
      return errors.New("spec.path cannot be updated")
    }
  }
  ```

6. Implement the controller, under normal circumstances this should be straightforward, just use this code:

  ```golang
    type MyVaultTypeReconciler struct {
      vaultresourcecontroller.ReconcilerBase
    }

    func (r *MyVaultTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
      _ = log.FromContext(ctx)

      // Fetch the instance
      instance := &redhatcopv1alpha1.MyVaultType{}
      err := r.GetClient().Get(ctx, req.NamespacedName, instance)
      if err != nil {
        if apierrors.IsNotFound(err) {
          // Request object not found, could have been deleted after reconcile request.
          // Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
          // Return and don't requeue
          return reconcile.Result{}, nil
        }
        // Error reading the object - requeue the request.
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

    func (r *MyVaultTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
      return ctrl.NewControllerManagedBy(mgr).
        For(&redhatcopv1alpha1.MyVaultType{},builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
        Complete(r)
    }
  ```

7. On the `main.go` update the controller reconciler to use the new operator util structur.

  ```golang
	if err = (&controllers.MyVaultTypeReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "MyVaultType")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MyVaultType")
		os.Exit(1)
	}  
  ...

  if webhooks, ok := os.LookupEnv("ENABLE_WEBHOOKS"); !ok || webhooks != "false" {
  ...
  	if err = (&redhatcopv1alpha1.MyVaultType{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RandomSecret")
			os.Exit(1)
		}
  ```

8. Regenerate the CRD manifests, RBAC, and DeepCopy methods using the Makefile (this ensures the pinned `controller-gen` version is used):

   ```shell
   make generate manifests
   ```

9. Build and test:

   ```shell
   make build test
   ```
