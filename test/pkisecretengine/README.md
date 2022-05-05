# Vault and Cert Manager Integration

## Prerequisite

- Cert Manager installed on the Cluster. Follow the [instruction](https://cert-manager.io/docs/installation/operator-lifecycle-manager/).

## Configure Vault

1. Create a Policy Admin to manage PKI Secret Engine.

    `oc create -f pki-secret-engine-admin-policy.yaml -n vault-admin`

2. Create and Authorize the default SA to create a PKI Engine and request certificates.

    `oc create -f pki-secret-engine-kube-auth-role.yaml -n vault-admin`

3. Create a PKI Secret Engine

    `oc create -f pki-secret-engine.yaml -n test-vault-config-operator`

4. Generate Root Certificate

    `oc create -f pki-secret-engine-config.yaml -n test-vault-config-operator`

5. Configure the PKI Role

    `oc create -f pki-secret-engine-role.yaml -n test-vault-config-operator`


## Configure Cert Manager

1. Retrieve the default SA secret token

    ```
    oc describe sa default -n vault-admin
    Name:                default
    Namespace:           vault-admin
    Labels:              <none>
    Annotations:         <none>
    Image pull secrets:  default-dockercfg-zhr22
    Mountable secrets:   default-dockercfg-zhr22
                         default-token-5t2tq
    Tokens:              default-token-5t2tq
                         default-token-q7lnl
    Events:              <none>

    export DEFAULT_SECRET=default-token-5t2tq
    ```


2. Get the `CA_BUNDLE` with the correct Certificate Authority in base64.
    
    ```
    oc extract cm/openshift-service-ca.crt --to=/tmp

    export CA_BUNDLE=$(base64 -w0 /tmp/service-ca.crt)
    ```

3. Create a Vault Issuer.

    ```
    envsubst < pki-secret-engine-issuer-sample.yaml | oc create -f -

    oc get issuer
    NAME    READY   AGE
    vault   True    77s
    ```

3. Create a dummy Certificate

    ```
    oc create -f pki-secret-engine-certificate-sample.yaml
    
    oc get certificate

    NAME                       READY   SECRET                          AGE
    vault-admin-issuer-dummy   True    vault-admin-issuer-dummy-cert   97s
    ```