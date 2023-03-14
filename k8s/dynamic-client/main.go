package main

import (
	"encoding/json"
	"flag"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	memory "k8s.io/client-go/discovery/cached"
	"log"
	"path/filepath"
)

var DeploymentResourceGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
var PodResourceGVR = schema.GroupVersionResource{Version: "v1", Resource: "pods"}
var VirtualServiceResourceGVR = schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}
var Namespace = "default"

func main() {
	client, cfg, err := initDynamicClient()
	if err != nil {
		log.Println("initDynamicClient err:", err)
		return
	}
	// 创建deploy
	//_ = createDeploy(client)

	// update deploy
	//err = updateDeploy(client)

	// 查询deploy列表
	//err = deployList(client)

	// 查询pod列表
	//err = podList(client)

	// 删除deploy
	//err = deleteDeploy(client)

	// 修改istio vs规则
	//err = patchVirtualServiceLabel(client)

	// 通过yaml文件创建deploy
	err = createDeployByYaml(client, cfg)
	if err != nil {
		fmt.Println("operate err: ", err)
		return
	}
}

func initConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home:=homedir.HomeDir(); home != "" {
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func initDynamicClient() (dynamic.Interface, *rest.Config, error) {
	cfg, err := initConfig()
	if err != nil {
		return nil, nil, err
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, cfg, nil
}

func createDeploy(client dynamic.Interface) error {
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "nginx-deploy",
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "nginx",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "nginx",
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "nginx",
								"image": "nginx:1.17.1",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	result, err := client.Resource(DeploymentResourceGVR).Namespace(Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil { //已经存在，再次创建会报错(deployments.apps "nginx-deploy" already exists)，可以改用Apply方法
		log.Println("client.Resource().Namespace().Create() err:", err)
		return err
	}
	log.Println("Create deployment: ", result.GetName())
	return nil
}

func updateDeploy(client dynamic.Interface) error {
	result, err := client.Resource(DeploymentResourceGVR).Namespace(Namespace).Get(context.TODO(), "nginx-deploy", metav1.GetOptions{})
	if err != nil {
		log.Println("failed to get latest version of Deployment err: ", err)
		return err
	}

	// update replicas to 1
	if err := unstructured.SetNestedField(result.Object, int64(1), "spec", "replicas"); err != nil {
		log.Println("failed to set replica value err: ", err)
		return err
	}

	_, err = client.Resource(DeploymentResourceGVR).Namespace(Namespace).Update(context.TODO(), result, metav1.UpdateOptions{})
	if err != nil {
		log.Println("deploy update err: ", err)
		return err
	}
	return nil
}

func podList(client dynamic.Interface) error {
	unstructObj, err := client.Resource(PodResourceGVR).Namespace(Namespace).List(context.TODO(), metav1.ListOptions{Limit: 100, LabelSelector:"app=nginx"})
	if err != nil {
		return err
	}
	if unstructObj == nil {
		return nil
	}

	podList := &apiv1.PodList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), podList)
	if err != nil {
		return err
	}
	for _, d := range podList.Items {
		fmt.Println(d.Name, ", ", d.Namespace, ", ", string(d.Status.Phase))
	}
	return nil
}

func deployList(client dynamic.Interface) error {
	unstructObj, err := client.Resource(DeploymentResourceGVR).Namespace(Namespace).List(context.TODO(), metav1.ListOptions{Limit: 100})
	if err != nil {
		return err
	}
	if unstructObj == nil {
		return nil
	}

	deploymentList := &appsv1.DeploymentList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), deploymentList)
	if err != nil {
		return err
	}
	for _, d := range deploymentList.Items {
		fmt.Println(d.Name, ", ", d.Namespace, ", ", d.Status.Replicas)
	}
	return nil
}

func deleteDeploy(client dynamic.Interface) error {
	err := client.Resource(DeploymentResourceGVR).Namespace(Namespace).Delete(context.TODO(), "nginx-deploy", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

type PatchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
/**
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: order-vs
  namespace: default
spec:
  hosts:
  - order-service
  http:
  - match:
    - headers:
        X-Canary-Label:
          exact: canary
    - sourceLabels:
        version: canary
    route:
      - destination:
          host: order-service
          subset: v1
  - route:
    - destination:
        host: order-service
        port:
          number: 50050
        subset: v1
      weight: 10
    - destination:
        host: order-service
        port:
          number: 50050
        subset: v2
      weight: 90
*/
func patchVirtualServiceLabel(client dynamic.Interface) error {
	virtualServiceName := ""
	label := "stable"
	patchPayload := make([]PatchStringValue, 1)
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/http/0/match/0/headers/X-Canary-Label/exact"
	patchPayload[0].Value = label
	patchPayload[1].Op = "replace"
	patchPayload[1].Path = "/spec/http/0/match/1/sourceLabels/version"
	patchPayload[1].Value = label
	patchBytes, err := json.Marshal(patchPayload)
	if err != nil {
		return err
	}

	_, err = client.Resource(VirtualServiceResourceGVR).Namespace(Namespace).Patch(context.TODO(), virtualServiceName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createDeployByYaml(client dynamic.Interface, cfg *rest.Config) error {
	var deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.17.1
`
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	obj := &unstructured.Unstructured{}
	_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode([]byte(deploymentYAML), nil, obj)
	if err != nil {
		return err
	}
	resourceMapper, err := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)).RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	var dClient dynamic.ResourceInterface
	if resourceMapper.Scope.Name() == meta.RESTScopeNameNamespace {
		dClient = client.Resource(resourceMapper.Resource).Namespace(obj.GetNamespace())
	} else {
		dClient = client.Resource(resourceMapper.Resource)
	}

	// 创建
	result, err := dClient.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(result.GetName())
	return nil
}