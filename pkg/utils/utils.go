package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/TwiN/go-color"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"sort"
	"strings"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ReturnManifestInFolder(folder string) []string {
	var manifests []string
	manifestFolder, _ := os.Open(folder)
	manifestFolderFiles, _ := manifestFolder.ReadDir(0)
	for _, file := range manifestFolderFiles {
		if strings.Contains(file.Name(), ".yaml") {
			manifests = append(manifests, folder+"/"+file.Name())
		}
	}
	// Sort alphebetically
	sort.Strings(manifests)
	return manifests
}

func KubeClientApply(file string, client *dynamic.DynamicClient, dClient *discovery.DiscoveryClient) ([]unstructured.Unstructured, error) {
	var appliedObjects []unstructured.Unstructured
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return appliedObjects, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dClient))
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(fileBytes), 100)
	for {

		var rawObj runtime.RawExtension
		err = decoder.Decode(&rawObj)
		if err != nil {
			if err.Error() == "EOF" {
				log.Printf(color.InBlue("[INFO] ")+"Reached end of file %s\n", file)
				break
			}
			return appliedObjects, err
		}
		obj := &unstructured.Unstructured{}
		_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, obj)
		if err != nil {
			return appliedObjects, err
		}
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return appliedObjects, err
		}
		log.Printf(color.InBlue("[INFO] ")+"Creating %s named %s\n", obj.GetKind(), obj.GetName())

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			if obj.GetNamespace() == "" {
				return appliedObjects, errors.New("Object should specify namespace, but namespace was empty")
			}
			dr = client.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = client.Resource(mapping.Resource)
		}
		data, err := json.Marshal(obj)
		if err != nil {
			return appliedObjects, err
		}
		_, err = dr.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: "cluster0-operators"})
		if err != nil {
			return appliedObjects, err
		}
		appliedObjects = append(appliedObjects, *obj)

	}
	return appliedObjects, nil
}

func NewKubeClients(kubeconfig string) (*dynamic.DynamicClient, *discovery.DiscoveryClient, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, err
		}
	}
	client, err := dynamic.NewForConfig(config)
	//client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	dClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return client, dClient, err
}
