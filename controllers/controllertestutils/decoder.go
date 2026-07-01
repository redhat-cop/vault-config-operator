package controllertestutils

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type decoder struct {
}

var runtimeDecoder runtime.Decoder

var errDecode = errors.New("failed to decode")

func init() {
	scheme := runtime.NewScheme()
	if err := redhatcopv1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	runtimeDecoder = serializer.NewCodecFactory(scheme).UniversalDeserializer()
}

func NewDecoder() *decoder {
	return new(decoder)
}

// CreateFromYAML reads a YAML fixture into an unstructured object (preserving only
// YAML-present fields), sets the namespace, and creates it via the API server.
// This allows CRD server-side defaulting to apply for absent fields.
func (d *decoder) CreateFromYAML(ctx context.Context, c client.Client, filename string, namespace string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	obj := &unstructured.Unstructured{}
	if err := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096).Decode(obj); err != nil {
		return "", err
	}

	obj.SetNamespace(namespace)
	if err := c.Create(ctx, obj); err != nil {
		return "", err
	}

	return obj.GetName(), nil
}

func decodeFile(filename string) (runtime.Object, *schema.GroupVersionKind, error) {
	stream, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	return runtimeDecoder.Decode(stream, nil, nil)
}

func DecodeInstance[T runtime.Object](filename string) (T, error) {
	obj, gvk, err := decodeFile(filename)
	if err != nil {
		var zero T
		return zero, err
	}

	t := reflect.TypeOf(*new(T))
	if t == nil || t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		var zero T
		return zero, errDecode
	}
	if gvk.Kind != t.Elem().Name() {
		var zero T
		return zero, errDecode
	}

	typed, ok := obj.(T)
	if !ok {
		var zero T
		return zero, errDecode
	}
	return typed, nil
}
