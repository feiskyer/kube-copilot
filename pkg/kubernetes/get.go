/*
Copyright 2023 - Present, Pengfei Ni

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubernetes

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

// GetYaml gets the yaml of a resource.
func GetYaml(resource, name, namespace string) (string, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return "", err
	}

	// Create a new clientset which include all needed client APIs
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	dynamicclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}

	grs, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return "", err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(grs)
	gvks, err := mapper.KindsFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return "", err
	}

	if len(gvks) == 0 {
		return "", fmt.Errorf("no kind found for %s", resource)
	}

	gvk := gvks[0]
	mapping, err := restmapper.NewDiscoveryRESTMapper(grs).RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", err
	}

	var dri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if namespace == "" {
			namespace = "default"
		}
		dri = dynamicclient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		dri = dynamicclient.Resource(mapping.Resource)
	}

	res, err := dri.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(res.Object)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
