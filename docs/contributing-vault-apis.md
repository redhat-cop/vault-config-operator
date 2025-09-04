# Contributing a new Vault API

All vault APIs can be manipulated using the Vault logical client and essentially with three operations: `read`, `write` (corresponding to create and update) and `delete`.

A framework has been created to make it simple to add new types (essentially new secret engine configuration and role types and new authentication engine configuration and role types).

Here are the steps:

1. Create the CRD type and miutating a validating webhooks:

   ```shell
   operator-sdk create api --group redhatcop --version v1alpha1 --kind MyVaultType --resource --controller
   operator-sdk create webhook --group redhatcop --version v1alpha1 --kind MyVaultType --defaulting --programmatic-validation
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
