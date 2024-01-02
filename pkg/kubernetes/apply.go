package kubernetes

import (
	"bytes"
	"context"
	"io"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// GetKubeConfig gets kubeconfig.
func GetKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

// ApplyYaml applies the manifests into Kubernetes cluster.
func ApplyYaml(manifests string) error {
	config, err := GetKubeConfig()
	if err != nil {
		return err
	}

	// Create a new clientset which include all needed client APIs
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	dynamicclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	// Decode the yaml file into a Kubernetes object
	decode := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manifests)), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decode.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		obj, gvk, err := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return err
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}

		grs, err := restmapper.GetAPIGroupResources(clientset.Discovery())
		if err != nil {
			return err
		}

		mapping, err := restmapper.NewDiscoveryRESTMapper(grs).RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dri = dynamicclient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = dynamicclient.Resource(mapping.Resource)
		}

		if _, err := dri.Apply(context.Background(), unstructuredObj.GetName(), unstructuredObj, metav1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
			return err
		}
	}

	return nil
}
