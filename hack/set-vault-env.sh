#!/bin/bash
# Script to set environment variables on the Vault Config Operator deployment
oc set env deployment/vault-config-operator-controller-manager -n vault-config-operator \
  VAULT_ADDR=http://vault.vault.svc:8200 \
  VAULT_SKIP_VERIFY=true
