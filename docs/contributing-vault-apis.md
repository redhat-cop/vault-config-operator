# Contributing a new Vault API

All vault APIs can be manipulated using teh Vault logical client and essentially with three operations: `read`, `write` (corresponding to create and update) and `delete`.

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
    // Authentication is the kube aoth configuraiton to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

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

4. Implements the `apis.ConditionsAware` interface
  
   ```golang
   var _ apis.ConditionsAware = &MyVaultType{}
   ```

5. Add needed validation and defaulting to the webhook. Notice that all the object will need to add the finalizer in the default webhook and to prevent the path from being changed in the valitating webhooks:

  ```golang:
  func (r *MyVaultType) Default() {
    authenginemountlog.Info("default", "name", r.Name)
    if !controllerutil.ContainsFinalizer(r, GetFinalizer(r)) {
      controllerutil.AddFinalizer(r, GetFinalizer(r))
    }
  }

  func (r *MyVaultType) ValidateUpdate(old runtime.Object) error {
    authenginemountlog.Info("validate update", "name", r.Name)

    // the path cannot be updated
    if r.Spec.Path != old.(*MyVaultType).Spec.Path {
      return errors.New("spec.path cannot be updated")
    }
  }
  ```

6. Implement the controller, under normal circumstances this should be straightfornad, just use this code:

  ```golang
    func (r *MyVaultTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
      _ = log.FromContext(ctx)
      instance := &redhatcopv1alpha1.MyVaultType{}
      err := r.GetClient().Get(ctx, req.NamespacedName, instance)
      if err != nil {
        if apierrors.IsNotFound(err) {
          return reconcile.Result{}, nil
        }
        return reconcile.Result{}, err
      }

      ctx = context.WithValue(ctx, "kubeClient", r.GetClient())
      vaultClient, err := instance.Spec.Authentication.GetVaultClient(ctx, instance.Namespace)
      if err != nil {
        r.Log.Error(err, "unable to create vault client", "instance", instance)
        return r.ManageError(ctx, instance, err)
      }
      ctx = context.WithValue(ctx, "vaultClient", vaultClient)
      vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)

      return vaultResource.Reconcile(ctx, instance)
    }

    func (r *MyVaultTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
      return ctrl.NewControllerManagedBy(mgr).
        For(&redhatcopv1alpha1.MyVaultType{}).
        Complete(r)
    }
  ```
