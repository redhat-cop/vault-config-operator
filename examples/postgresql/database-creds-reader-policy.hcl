# Configure read secrets
path "/{{identity.entity.aliases.auth_kubernetes_05e62199.metadata.service_account_namespace}}/database/creds/+" {
  capabilities = ["read"]
}