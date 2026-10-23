package utils

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestIdentityCodec(t *testing.T) {
	type Simple struct {
		KeyType string `vault:"key_type"`
	}
	result := StructToVaultPayload(&Simple{KeyType: "otp"})
	if result["key_type"] != "otp" {
		t.Errorf("expected 'otp', got %v", result["key_type"])
	}
}

func TestSkipField(t *testing.T) {
	type WithSkip struct {
		Name string `vault:"name"`
		Conn string `vault:"-"`
	}
	result := StructToVaultPayload(&WithSkip{Name: "test", Conn: "skip-me"})
	if _, exists := result["Conn"]; exists {
		t.Error("expected Conn to be skipped")
	}
	if _, exists := result["name"]; !exists {
		t.Error("expected name to be present")
	}
}

func TestDurationCodec(t *testing.T) {
	type WithDuration struct {
		TTL string `vault:"token_ttl,duration"`
	}
	result := StructToVaultPayload(&WithDuration{TTL: "4h"})
	expected := json.Number("14400")
	if result["token_ttl"] != expected {
		t.Errorf("expected %v, got %v", expected, result["token_ttl"])
	}
}

func TestDurationCodecEmpty(t *testing.T) {
	type WithDuration struct {
		TTL string `vault:"token_ttl,duration"`
	}
	result := StructToVaultPayload(&WithDuration{TTL: ""})
	expected := json.Number("0")
	if result["token_ttl"] != expected {
		t.Errorf("expected %v, got %v", expected, result["token_ttl"])
	}
}

func TestNumberCodec(t *testing.T) {
	type WithNumber struct {
		Port int `vault:"port,number"`
	}
	result := StructToVaultPayload(&WithNumber{Port: 22})
	expected := json.Number("22")
	if result["port"] != expected {
		t.Errorf("expected %v, got %v", expected, result["port"])
	}
}

func TestIntmapCodec(t *testing.T) {
	type WithIntMap struct {
		Lengths map[string]int `vault:"allowed_user_key_lengths,intmap"`
	}
	result := StructToVaultPayload(&WithIntMap{Lengths: map[string]int{"rsa": 2048}})
	m, ok := result["allowed_user_key_lengths"].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if m["rsa"] != json.Number("2048") {
		t.Errorf("expected json.Number(2048), got %v", m["rsa"])
	}
}

func TestIntmapCodecNil(t *testing.T) {
	type WithIntMap struct {
		Lengths map[string]int `vault:"allowed_user_key_lengths,intmap"`
	}
	result := StructToVaultPayload(&WithIntMap{})
	m, ok := result["allowed_user_key_lengths"].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestStringmapCodec(t *testing.T) {
	type WithStringMap struct {
		Extensions map[string]string `vault:"default_extensions,stringmap"`
	}
	result := StructToVaultPayload(&WithStringMap{Extensions: map[string]string{"permit-pty": ""}})
	m, ok := result["default_extensions"].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if m["permit-pty"] != "" {
		t.Errorf("expected empty string, got %v", m["permit-pty"])
	}
}

func TestStringmapCodecNil(t *testing.T) {
	type WithStringMap struct {
		Extensions map[string]string `vault:"default_extensions,stringmap"`
	}
	result := StructToVaultPayload(&WithStringMap{})
	m, ok := result["default_extensions"].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestWhenConditionalIncluded(t *testing.T) {
	type WithWhen struct {
		KeyType  string `vault:"key_type"`
		CIDRList string `vault:"cidr_list,when=KeyType:otp"`
	}
	result := StructToVaultPayload(&WithWhen{KeyType: "otp", CIDRList: "10.0.0.0/8"})
	if result["cidr_list"] != "10.0.0.0/8" {
		t.Errorf("expected cidr_list when KeyType=otp, got %v", result["cidr_list"])
	}
}

func TestWhenConditionalExcluded(t *testing.T) {
	type WithWhen struct {
		KeyType  string `vault:"key_type"`
		CIDRList string `vault:"cidr_list,when=KeyType:otp"`
	}
	result := StructToVaultPayload(&WithWhen{KeyType: "ca", CIDRList: "10.0.0.0/8"})
	if _, exists := result["cidr_list"]; exists {
		t.Error("expected cidr_list to be excluded when KeyType=ca")
	}
}

func TestDefaultBoolNil(t *testing.T) {
	type WithDefault struct {
		Verify *bool `vault:"verify_connection,default=true"`
	}
	result := StructToVaultPayload(&WithDefault{Verify: nil})
	if result["verify_connection"] != true {
		t.Errorf("expected true default, got %v", result["verify_connection"])
	}
}

func TestDefaultBoolNonNil(t *testing.T) {
	type WithDefault struct {
		Verify *bool `vault:"verify_connection,default=true"`
	}
	f := false
	result := StructToVaultPayload(&WithDefault{Verify: &f})
	if result["verify_connection"] != false {
		t.Errorf("expected false, got %v", result["verify_connection"])
	}
}

func TestDefaultIntNil(t *testing.T) {
	type WithDefault struct {
		MaxRetries *int `vault:"max_retries,number,default=-1"`
	}
	result := StructToVaultPayload(&WithDefault{MaxRetries: nil})
	if result["max_retries"] != json.Number("-1") {
		t.Errorf("expected json.Number(-1), got %v", result["max_retries"])
	}
}

func TestSpreadmap(t *testing.T) {
	type WithSpread struct {
		Name   string            `vault:"name"`
		Config map[string]string `vault:",spreadmap"`
	}
	result := StructToVaultPayload(&WithSpread{
		Name:   "test",
		Config: map[string]string{"host": "localhost", "port": "5432"},
	})
	if result["name"] != "test" {
		t.Errorf("expected 'test', got %v", result["name"])
	}
	if result["host"] != "localhost" {
		t.Errorf("expected 'localhost', got %v", result["host"])
	}
	if result["port"] != "5432" {
		t.Errorf("expected '5432', got %v", result["port"])
	}
}

func TestSpreadmapEmpty(t *testing.T) {
	type WithSpread struct {
		Name   string            `vault:"name"`
		Config map[string]string `vault:",spreadmap"`
	}
	result := StructToVaultPayload(&WithSpread{Name: "test"})
	if _, exists := result["Config"]; exists {
		t.Error("empty spreadmap should not add entries")
	}
	if result["name"] != "test" {
		t.Errorf("expected 'test', got %v", result["name"])
	}
}

func TestInline(t *testing.T) {
	type Inner struct {
		KeyType string `vault:"key_type"`
		TTL     string `vault:"ttl,duration"`
	}
	type Outer struct {
		Name  string `vault:"name"`
		Inner Inner  `vault:",inline"`
	}
	result := StructToVaultPayload(&Outer{
		Name:  "test",
		Inner: Inner{KeyType: "otp", TTL: "1h"},
	})
	if result["name"] != "test" {
		t.Errorf("expected 'test', got %v", result["name"])
	}
	if result["key_type"] != "otp" {
		t.Errorf("expected 'otp', got %v", result["key_type"])
	}
	if result["ttl"] != json.Number("3600") {
		t.Errorf("expected 3600, got %v", result["ttl"])
	}
}

func TestOmitemptyString(t *testing.T) {
	type WithOmit struct {
		Name    string `vault:"name"`
		Comment string `vault:"comment,omitempty"`
	}
	result := StructToVaultPayload(&WithOmit{Name: "test"})
	if _, exists := result["comment"]; exists {
		t.Error("empty string with omitempty should be excluded")
	}
}

func TestOmitemptyStringNonEmpty(t *testing.T) {
	type WithOmit struct {
		Name    string `vault:"name"`
		Comment string `vault:"comment,omitempty"`
	}
	result := StructToVaultPayload(&WithOmit{Name: "test", Comment: "hello"})
	if result["comment"] != "hello" {
		t.Errorf("expected 'hello', got %v", result["comment"])
	}
}

func TestOmitemptyPtrNil(t *testing.T) {
	type WithOmit struct {
		Password *string `vault:"password,omitempty"`
	}
	result := StructToVaultPayload(&WithOmit{})
	if _, exists := result["password"]; exists {
		t.Error("nil pointer with omitempty should be excluded")
	}
}

func TestSliceToAny(t *testing.T) {
	type WithSlice struct {
		Roles []string `vault:"allowed_roles"`
	}
	result := StructToVaultPayload(&WithSlice{Roles: []string{"a", "b"}})
	expected := []any{"a", "b"}
	if !reflect.DeepEqual(result["allowed_roles"], expected) {
		t.Errorf("expected %v, got %v", expected, result["allowed_roles"])
	}
}

func TestUntaggedFieldsIgnored(t *testing.T) {
	type Mixed struct {
		Tagged   string `vault:"tagged"`
		Untagged string
		JsonOnly string `json:"jsonOnly"`
	}
	result := StructToVaultPayload(&Mixed{Tagged: "yes", Untagged: "no", JsonOnly: "no"})
	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d: %v", len(result), result)
	}
}

func TestBoolField(t *testing.T) {
	type WithBool struct {
		AllowCerts bool `vault:"allow_user_certificates"`
	}
	result := StructToVaultPayload(&WithBool{AllowCerts: true})
	if result["allow_user_certificates"] != true {
		t.Errorf("expected true, got %v", result["allow_user_certificates"])
	}
}

func TestBoolFieldFalse(t *testing.T) {
	type WithBool struct {
		AllowCerts bool `vault:"allow_user_certificates"`
	}
	result := StructToVaultPayload(&WithBool{AllowCerts: false})
	if result["allow_user_certificates"] != false {
		t.Errorf("expected false, got %v", result["allow_user_certificates"])
	}
}
