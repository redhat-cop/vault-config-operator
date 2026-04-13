package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRabbitMQSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *RabbitMQSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &RabbitMQSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: RabbitMQSecretEngineRoleSpec{
					Path: "rabbitmq",
					Name: "custom-name",
				},
			},
			expectedPath: vaultutils.CleansePath("rabbitmq/" + "roles" + "/" + "custom-name"),
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &RabbitMQSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: RabbitMQSecretEngineRoleSpec{
					Path: "rabbitmq",
				},
			},
			expectedPath: vaultutils.CleansePath("rabbitmq/" + "roles" + "/" + "meta-name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestRMQSERoleRabbitMQToMap(t *testing.T) {
	vhosts := []Vhost{{
		VhostName: "/",
		Permissions: VhostPermissions{
			Configure: ".*",
			Write:     ".*",
			Read:      ".*",
		},
	}}
	vhostTopics := []VhostTopic{{
		VhostName: "/",
		Topics: []Topic{{
			TopicName: "amq.topic",
			Permissions: VhostPermissions{
				Configure: "",
				Write:     ".*",
				Read:      ".*",
			},
		}},
	}}

	role := RMQSERole{
		Tags:        "administrator",
		Vhosts:      vhosts,
		VhostTopics: vhostTopics,
	}

	result := role.rabbitMQToMap()

	expectedKeys := []string{"tags", "vhosts", "vhost_topics"}
	if len(result) != len(expectedKeys) {
		t.Errorf("rabbitMQToMap() len = %d, want %d keys", len(result), len(expectedKeys))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in rabbitMQToMap() output", key)
		}
	}

	if result["tags"] != "administrator" {
		t.Errorf("tags = %v", result["tags"])
	}

	vhostsStr, ok := result["vhosts"].(string)
	if !ok {
		t.Fatalf("vhosts type = %T, want string", result["vhosts"])
	}
	wantVhosts := convertVhostsToJson(vhosts)
	if vhostsStr != wantVhosts {
		t.Errorf("vhosts = %q, want %q", vhostsStr, wantVhosts)
	}

	topicsStr, ok := result["vhost_topics"].(string)
	if !ok {
		t.Fatalf("vhost_topics type = %T, want string", result["vhost_topics"])
	}
	wantTopics := convertTopicsToJson(vhostTopics)
	if topicsStr != wantTopics {
		t.Errorf("vhost_topics = %q, want %q", topicsStr, wantTopics)
	}
}

func TestRabbitMQSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &RabbitMQSecretEngineRole{
		Spec: RabbitMQSecretEngineRoleSpec{
			Path: "rabbitmq",
			RMQSERole: RMQSERole{
				Tags: "administrator",
				Vhosts: []Vhost{{
					VhostName: "/",
					Permissions: VhostPermissions{
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				}},
				VhostTopics: []VhostTopic{{
					VhostName: "/",
					Topics: []Topic{{
						TopicName: "amq.topic",
						Permissions: VhostPermissions{
							Configure: "",
							Write:     ".*",
							Read:      ".*",
						},
					}},
				}},
			},
		},
	}

	payload := role.Spec.RMQSERole.rabbitMQToMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestRabbitMQSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &RabbitMQSecretEngineRole{
		Spec: RabbitMQSecretEngineRoleSpec{
			Path: "rabbitmq",
			RMQSERole: RMQSERole{
				Tags: "administrator",
				Vhosts: []Vhost{{
					VhostName: "/",
					Permissions: VhostPermissions{
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				}},
			},
		},
	}

	payload := role.Spec.RMQSERole.rabbitMQToMap()
	payload["tags"] = "management"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload to NOT be equivalent")
	}
}

func TestRabbitMQSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &RabbitMQSecretEngineRole{
		Spec: RabbitMQSecretEngineRoleSpec{
			Path: "rabbitmq",
			RMQSERole: RMQSERole{
				Tags: "administrator",
				Vhosts: []Vhost{{
					VhostName: "/",
					Permissions: VhostPermissions{
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				}},
			},
		},
	}

	payload := role.Spec.RMQSERole.rabbitMQToMap()
	payload["extra_vault_field"] = "x"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent")
	}
}

func TestRabbitMQSecretEngineRoleIsDeletable(t *testing.T) {
	role := &RabbitMQSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected RabbitMQSecretEngineRole to be deletable")
	}
}

func TestRabbitMQSecretEngineRoleConditions(t *testing.T) {
	role := &RabbitMQSecretEngineRole{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	role.SetConditions(conditions)
	got := role.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("expected condition status True, got %v", got[0].Status)
	}
}
