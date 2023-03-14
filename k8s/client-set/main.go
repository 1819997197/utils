package main

import (
	"flag"
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/pointer"
	"log"
	"path/filepath"
)

const (
	NAMESPACE = "default"
)

func main() {
	client, err := initClientSet()
	if err != nil {
		log.Println("initClientSet err: ", err)
		return
	}
	// 创建deploy
	createDeploy(client)

	// 创建service
	createService(client)

	// 删除deploy、service
	err = deleteAll(client)
	if err != nil {
		log.Println("deleteAll err: ", err)
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

func initClientSet() (*kubernetes.Clientset, error) {
	cfg, err := initConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func createDeploy(clientSet *kubernetes.Clientset) {
	deployClient := clientSet.AppsV1().Deployments(NAMESPACE) //通过该客户端可以进行CURD操作
	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deploy",
			Namespace: NAMESPACE,
		},
		Spec: v1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.17.1",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
		},
		Status: v1.DeploymentStatus{},
	}
	// 创建
	result, err := deployClient.Create(context.TODO(), deploy, metav1.CreateOptions{})
	if err != nil {
		log.Println("deployClient.Create err:", err)
		panic(err)
	}
	fmt.Println(result.GetName())
}

func createService(clientSet *kubernetes.Clientset) {
	serviceClient := clientSet.CoreV1().Services(NAMESPACE)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-service",
			Namespace: NAMESPACE,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					NodePort: 30001,
				},
			},
			Selector: map[string]string{
				"app": "nginx",
			},
			Type: corev1.ServiceTypeNodePort,
		},
		Status: corev1.ServiceStatus{},
	}

	result, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		log.Println("serviceClient.Create err:", err)
		panic(err)
	}
	fmt.Println(result.GetName())
}

func deleteAll(clientSet *kubernetes.Clientset) error {
	emptyDeleteOptions := metav1.DeleteOptions{}
	err := clientSet.CoreV1().Services(NAMESPACE).Delete(context.TODO(), "nginx-service", emptyDeleteOptions)
	if err != nil {
		return err //service 不存在时会报错(services "nginx-service" not found)
	}

	err = clientSet.AppsV1().Deployments(NAMESPACE).Delete(context.TODO(), "nginx-deploy", emptyDeleteOptions)
	if err != nil {
		return err //deploy 不存在时会报错(deployments.apps "nginx-deploy" not found)
	}
	return nil
}