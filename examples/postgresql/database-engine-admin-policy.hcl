# mount database secrets engines
path "/sys/mounts/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
  allowed_parameters = {
    "type" = ["database"]
    "*"   = []
  }
}

# Configure database secrets engines
path "/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database/config/+" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# Configure database roles
path "/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database/roles/+" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}