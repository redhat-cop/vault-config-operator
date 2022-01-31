# Vault and Cert Manager Integration

## Prerequisite

- Cert Manager installed on the Cluster. Follow the [instruction](https://cert-manager.io/docs/installation/operator-lifecycle-manager/).

## Configure Vault

1. Create a PKI Secret Engine

    `oc create -f pki-secret-engine.yaml`

2. Generate Root Certificate

    `oc create -f pki-secret-engine-config.yaml`

3. Configure the PKI Role

    `oc create -f pki-secret-engine-role.yaml`

4. Define a Vault Policy

    `oc create -f pki-secret-engine-policy.yaml`

5. Authorize the default SA to request certificates.

    `oc create -f pki-secret-engine-kube-auth-role.yaml`

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