package controllertestutils

import (
	"errors"
	"io/ioutil"
	"reflect"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type decoder struct {
}

var runtimeDecoder runtime.Decoder

var errDecode = errors.New("failed to decode")

func init() {
	scheme := runtime.NewScheme()
	redhatcopv1alpha1.AddToScheme(scheme)
	runtimeDecoder = serializer.NewCodecFactory(scheme).UniversalDeserializer()
}

func NewDecoder() *decoder {
	return new(decoder)
}

func (d *decoder) decodeFile(filename string) (runtime.Object, *schema.GroupVersionKind, error) {
	stream, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	return runtimeDecoder.Decode(stream, nil, nil)
}

func (d *decoder) GetPasswordPolicyInstance(filename string) (*redhatcopv1alpha1.PasswordPolicy, error) {
	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.PasswordPolicy{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.PasswordPolicy)
		return o, nil
	}

	return nil, errDecode
}

func (d *decoder) GetVaultSecretInstance(filename string) (*redhatcopv1alpha1.VaultSecret, error) {

	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.VaultSecret{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.VaultSecret)
		return o, nil
	}

	return nil, errDecode
}

func (d *decoder) GetPolicyInstance(filename string) (*redhatcopv1alpha1.Policy, error) {
	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.Policy{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.Policy)
		return o, nil
	}

	return nil, errDecode
}

func (d *decoder) GetKubernetesAuthEngineRoleInstance(filename string) (*redhatcopv1alpha1.KubernetesAuthEngineRole, error) {
	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.KubernetesAuthEngineRole{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.KubernetesAuthEngineRole)
		return o, nil
	}

	return nil, errDecode
}

func (d *decoder) GetSecretEngineMountInstance(filename string) (*redhatcopv1alpha1.SecretEngineMount, error) {
	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.SecretEngineMount{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.SecretEngineMount)
		return o, nil
	}

	return nil, errDecode
}

func (d *decoder) GetRandomSecretInstance(filename string) (*redhatcopv1alpha1.RandomSecret, error) {
	obj, groupKindVersion, err := d.decodeFile(filename)
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.RandomSecret{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.RandomSecret)
		return o, nil
	}

	return nil, errDecode
}
