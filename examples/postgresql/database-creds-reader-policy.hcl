# Configure read secrets
path "/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database/creds/+" {
  capabilities = ["read"]
}