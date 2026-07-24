package v1alpha1

import (
	"encoding/json"
	"reflect"
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
	wantVhosts := `{"/":{"configure":".*","write":".*","read":".*"}}`
	if vhostsStr != wantVhosts {
		t.Errorf("vhosts = %q, want %q", vhostsStr, wantVhosts)
	}

	topicsStr, ok := result["vhost_topics"].(string)
	if !ok {
		t.Fatalf("vhost_topics type = %T, want string", result["vhost_topics"])
	}
	wantTopics := `{"/":{"amq.topic":{"write":".*","read":".*"}}}`
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

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
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

func TestConvertVhostsToJsonMultipleEntries(t *testing.T) {
	vhosts := []Vhost{
		{
			VhostName: "/",
			Permissions: VhostPermissions{
				Configure: ".*",
				Write:     ".*",
				Read:      ".*",
			},
		},
		{
			VhostName: "/staging",
			Permissions: VhostPermissions{
				Configure: "",
				Write:     "staging-.*",
				Read:      ".*",
			},
		},
		{
			VhostName: "/production",
			Permissions: VhostPermissions{
				Configure: "",
				Write:     "",
				Read:      "prod-.*",
			},
		},
	}

	result := convertVhostsToJson(vhosts)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("expected 3 vhost entries, got %d", len(parsed))
	}

	for _, name := range []string{"/", "/staging", "/production"} {
		if _, ok := parsed[name]; !ok {
			t.Errorf("expected vhost %q to be present in result", name)
		}
	}

	rootPerms, ok := parsed["/"].(map[string]any)
	if !ok {
		t.Fatalf("expected / permissions to be a map, got %T", parsed["/"])
	}
	if rootPerms["configure"] != ".*" || rootPerms["write"] != ".*" || rootPerms["read"] != ".*" {
		t.Errorf("unexpected / permissions: %v", rootPerms)
	}

	stagingPerms, ok := parsed["/staging"].(map[string]any)
	if !ok {
		t.Fatalf("expected /staging permissions to be a map, got %T", parsed["/staging"])
	}
	if stagingPerms["write"] != "staging-.*" || stagingPerms["read"] != ".*" {
		t.Errorf("unexpected /staging permissions: %v", stagingPerms)
	}
}

func TestConvertTopicsToJsonMultipleEntries(t *testing.T) {
	vhosts := []VhostTopic{
		{
			VhostName: "/",
			Topics: []Topic{
				{
					TopicName: "amq.topic",
					Permissions: VhostPermissions{
						Write: ".*",
						Read:  ".*",
					},
				},
				{
					TopicName: "amq.fanout",
					Permissions: VhostPermissions{
						Write: "fanout-.*",
						Read:  ".*",
					},
				},
			},
		},
		{
			VhostName: "/staging",
			Topics: []Topic{
				{
					TopicName: "amq.direct",
					Permissions: VhostPermissions{
						Write: "staging-.*",
						Read:  "staging-.*",
					},
				},
			},
		},
	}

	result := convertTopicsToJson(vhosts)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 vhost entries, got %d", len(parsed))
	}

	rootTopics, ok := parsed["/"].(map[string]any)
	if !ok {
		t.Fatalf("expected / to be a map, got %T", parsed["/"])
	}
	if len(rootTopics) != 2 {
		t.Errorf("expected 2 topics under /, got %d", len(rootTopics))
	}
	if _, ok := rootTopics["amq.topic"]; !ok {
		t.Error("expected amq.topic under /")
	}
	if _, ok := rootTopics["amq.fanout"]; !ok {
		t.Error("expected amq.fanout under /")
	}

	fanoutPerms, ok := rootTopics["amq.fanout"].(map[string]any)
	if !ok {
		t.Fatalf("expected amq.fanout to be a map, got %T", rootTopics["amq.fanout"])
	}
	if fanoutPerms["write"] != "fanout-.*" {
		t.Errorf("expected amq.fanout write = fanout-.*, got %v", fanoutPerms["write"])
	}

	stagingTopics, ok := parsed["/staging"].(map[string]any)
	if !ok {
		t.Fatalf("expected /staging to be a map, got %T", parsed["/staging"])
	}
	if len(stagingTopics) != 1 {
		t.Errorf("expected 1 topic under /staging, got %d", len(stagingTopics))
	}
	if _, ok := stagingTopics["amq.direct"]; !ok {
		t.Error("expected amq.direct under /staging")
	}
}

func TestRMQSERoleRabbitMQToMapMultipleVhosts(t *testing.T) {
	role := RMQSERole{
		Tags: "administrator",
		Vhosts: []Vhost{
			{
				VhostName:   "/",
				Permissions: VhostPermissions{Configure: ".*", Write: ".*", Read: ".*"},
			},
			{
				VhostName:   "/staging",
				Permissions: VhostPermissions{Configure: "", Write: "staging-.*", Read: ".*"},
			},
		},
		VhostTopics: []VhostTopic{
			{
				VhostName: "/",
				Topics: []Topic{
					{TopicName: "amq.topic", Permissions: VhostPermissions{Write: ".*", Read: ".*"}},
					{TopicName: "amq.fanout", Permissions: VhostPermissions{Write: "fanout-.*", Read: ".*"}},
				},
			},
			{
				VhostName: "/staging",
				Topics: []Topic{
					{TopicName: "amq.direct", Permissions: VhostPermissions{Write: "staging-.*", Read: "staging-.*"}},
				},
			},
		},
	}

	result := role.rabbitMQToMap()

	if result["tags"] != "administrator" {
		t.Errorf("tags = %v, want administrator", result["tags"])
	}

	vhostsStr, ok := result["vhosts"].(string)
	if !ok {
		t.Fatalf("vhosts type = %T, want string", result["vhosts"])
	}
	var actualVhosts, expectedVhosts map[string]any
	if err := json.Unmarshal([]byte(vhostsStr), &actualVhosts); err != nil {
		t.Fatalf("failed to unmarshal actual vhosts: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"/":{"configure":".*","write":".*","read":".*"},"/staging":{"write":"staging-.*","read":".*"}}`), &expectedVhosts); err != nil {
		t.Fatalf("failed to unmarshal expected vhosts: %v", err)
	}
	if !reflect.DeepEqual(actualVhosts, expectedVhosts) {
		t.Errorf("vhosts mismatch\ngot:  %v\nwant: %v", actualVhosts, expectedVhosts)
	}

	topicsStr, ok := result["vhost_topics"].(string)
	if !ok {
		t.Fatalf("vhost_topics type = %T, want string", result["vhost_topics"])
	}
	var actualTopics, expectedTopics map[string]any
	if err := json.Unmarshal([]byte(topicsStr), &actualTopics); err != nil {
		t.Fatalf("failed to unmarshal actual topics: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"/":{"amq.fanout":{"write":"fanout-.*","read":".*"},"amq.topic":{"write":".*","read":".*"}},"/staging":{"amq.direct":{"write":"staging-.*","read":"staging-.*"}}}`), &expectedTopics); err != nil {
		t.Fatalf("failed to unmarshal expected topics: %v", err)
	}
	if !reflect.DeepEqual(actualTopics, expectedTopics) {
		t.Errorf("vhost_topics mismatch\ngot:  %v\nwant: %v", actualTopics, expectedTopics)
	}
}

func TestRabbitMQSecretEngineRoleIsEquivalentMultipleVhosts(t *testing.T) {
	role := &RabbitMQSecretEngineRole{
		Spec: RabbitMQSecretEngineRoleSpec{
			Path: "rabbitmq",
			RMQSERole: RMQSERole{
				Tags: "administrator",
				Vhosts: []Vhost{
					{VhostName: "/", Permissions: VhostPermissions{Configure: ".*", Write: ".*", Read: ".*"}},
					{VhostName: "/staging", Permissions: VhostPermissions{Configure: "", Write: "staging-.*", Read: ".*"}},
				},
				VhostTopics: []VhostTopic{
					{
						VhostName: "/",
						Topics: []Topic{
							{TopicName: "amq.topic", Permissions: VhostPermissions{Write: ".*", Read: ".*"}},
						},
					},
					{
						VhostName: "/staging",
						Topics: []Topic{
							{TopicName: "amq.direct", Permissions: VhostPermissions{Write: "staging-.*", Read: "staging-.*"}},
						},
					},
				},
			},
		},
	}

	matchingPayload := map[string]any{
		"tags":         "administrator",
		"vhosts":       `{"/":{"configure":".*","write":".*","read":".*"},"/staging":{"write":"staging-.*","read":".*"}}`,
		"vhost_topics": `{"/":{"amq.topic":{"write":".*","read":".*"}},"/staging":{"amq.direct":{"write":"staging-.*","read":"staging-.*"}}}`,
		"extra_field":  "should-be-ignored",
	}

	if !role.IsEquivalentToDesiredState(matchingPayload) {
		t.Error("expected matching multi-entry payload to be equivalent")
	}

	mismatchPayload := map[string]any{
		"tags":         "administrator",
		"vhosts":       `{"/":{"configure":".*","write":".*","read":".*"}}`,
		"vhost_topics": `{"/":{"amq.topic":{"write":".*","read":".*"}},"/staging":{"amq.direct":{"write":"staging-.*","read":"staging-.*"}}}`,
	}

	if role.IsEquivalentToDesiredState(mismatchPayload) {
		t.Error("expected payload missing /staging vhost to NOT be equivalent")
	}
}
