package vaultresourcecontroller

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetDynamicClientOnUnstructured returns a dynamic client on an Unstructured type. This client can be further namespaced.
// needs context with log and restConfig
// TODO consider refactoring using apimachinery.RESTClientForGVK in controller-runtime
func GetDynamicClientOnUnstructured(context context.Context, obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	log := log.FromContext(context)
	apiRes, err := getAPIReourceForGVK(context, obj.GetObjectKind().GroupVersionKind())
	if err != nil {
		log.Error(err, "Unable to get apiresource from unstructured", "unstructured", obj)
		return nil, err
	}
	dc, err := GetDynamicClientForAPIResource(context, apiRes)
	if err != nil {
		log.Error(err, "Unable to get namespaceable dynamic client from ", "resource", apiRes)
		return nil, err
	}
	if apiRes.Namespaced {
		return dc.Namespace(obj.GetNamespace()), nil
	}
	return dc, nil
}

// GetDynamicClientOnAPIResource returns a dynamic client on an APIResource. This client can be further namespaced.
// needs context with log and restConfig
func GetDynamicClientForAPIResource(context context.Context, resource *metav1.APIResource) (dynamic.NamespaceableResourceInterface, error) {
	return getDynamicClientForGVR(context, schema.GroupVersionResource{
		Group:    resource.Group,
		Version:  resource.Version,
		Resource: resource.Name,
	})
}

func getDynamicClientForGVR(context context.Context, gvr schema.GroupVersionResource) (dynamic.NamespaceableResourceInterface, error) {
	log := log.FromContext(context)
	restConfig := context.Value("restConfig").(*rest.Config)
	intf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		log.Error(err, "Unable to get dynamic client")
		return nil, err
	}
	res := intf.Resource(gvr)
	return res, nil
}

// GetDynamicClientForGVK returns a dynamic client on an gvk type. Also returns whether this reosurce is namespaced. This client can be further namespaced.
// needs context with log and restConfig
func GetDynamicClientForGVK(context context.Context, gvk schema.GroupVersionKind) (dynamic.NamespaceableResourceInterface, bool, error) {
	log := log.FromContext(context)
	apiRes, err := getAPIReourceForGVK(context, gvk)
	if err != nil {
		log.Error(err, "unable to get apiresource from", "gvk", gvk)
		return nil, false, err
	}
	nri, err := GetDynamicClientForAPIResource(context, apiRes)
	if err != nil {
		log.Error(err, "unable to get dynamic client from", "apires", apiRes)
		return nil, false, err
	}
	return nri, apiRes.Namespaced, nil
}

func getAPIReourceForGVK(context context.Context, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	res := &metav1.APIResource{}
	log := log.FromContext(context)
	restConfig := context.Value("restConfig").(*rest.Config)
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(restConfig)
	resList, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		log.Error(err, "unable to retrieve resource list for", "gvk", gvk.GroupVersion().String())
		return nil, err
	}
	for i := range resList.APIResources {
		//if a resource contains a "/" it's referencing a subresource. we don't support subresource for now.
		if resList.APIResources[i].Kind == gvk.Kind && !strings.Contains(resList.APIResources[i].Name, "/") {
			res = &resList.APIResources[i]
			res.Group = gvk.Group
			res.Version = gvk.Version
			break
		}
	}
	return res, nil
}

// SetIndexField this function allows to prepare an index field for an objct so that fieldSelector can be used.
// It needs a cache object probably obtained via mrg.GetCache()
// This is a generic implementation, so it's relatively slow
// path should be expressed in the form of .<field>.<field> ...
// needs context with log
func SetIndexField(context context.Context, cache cache.Cache, obj client.Object, path string) error {
	log := log.FromContext(context)
	return cache.IndexField(context, obj, path, func(o client.Object) []string {
		mapObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
		if err != nil {
			log.Error(err, "unable to convert object to unstructured ", "object", o)
			return nil
		}
		jp := jsonpath.New("fieldPath:" + path)
		err = jp.Parse("{ $" + path + "}")
		if err != nil {
			log.Error(err, "unable to parse ", "fieldPath", path)
			return nil
		}
		values, err := jp.FindResults(mapObj)
		if err != nil {
			log.Error(err, "unable to apply ", "jsonpath", jp, " to obj ", mapObj)
			return nil
		}
		result := []string{}
		if len(values) > 0 {
			for i := range values[0] {
				if val, ok := values[0][i].Interface().(string); ok {
					result = append(result, val)
				}
			}
		}
		return result
	})
}
