package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	watch "k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Set the path to your kubeconfig file

	homeDir,_:=os.UserHomeDir()

	kubeconfigPath:=filepath.Join(homeDir, ".kube", "config")

	// Build the configuration from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create a new clientset using the configuration
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	namespace:="default"

	// pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
    // if err != nil {
    //         fmt.Printf("error getting pods: %v\n", err)
    // }

	// fmt.Println(pods.Size())

	// Define the deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(4),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "example-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "example-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "example-container",
						Image: "nginx:latest",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
					}},
					SchedulingGates: []corev1.PodSchedulingGate{corev1.PodSchedulingGate{Name:"sg1"},corev1.PodSchedulingGate{Name:"sg2"},},
				},
			},
		},
	}

	// Create the deployment
	result, err := clientset.AppsV1().Deployments(namespace).Create(context.TODO(),deployment,metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Failed to create deployment: %v", err)
	}

	fmt.Printf("Deployment created: %s\n", result.GetObjectMeta().GetName())


	// watch state changes

	watcher, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal("Error creating watcher:", err.Error())
	}

	go func(){
		for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod := event.Object.(*corev1.Pod)
			fmt.Println("Pod added:", pod.Name)
			// Handle added pod event here
			printPod(pod)

		case watch.Modified:
			pod := event.Object.(*corev1.Pod)
			fmt.Println("Pod modified:", pod.Name)
			// Handle modified pod event here
			printPod(pod)
			

		case watch.Deleted:
			pod := event.Object.(*corev1.Pod)
			fmt.Println("Pod deleted:", pod.Name)
			// Handle deleted pod event here
			printPod(pod)

		case watch.Error:
			err := event.Object.(*metav1.Status)
			log.Println("Error watching pods:", err.Message)
			// Handle error event here
		}

		}
	}()


	
	
	//remove

	// kubectl patch pod example-deployment-6876b95d78-crc2q --type=json -p='[{"op": "replace","path": "/spec/schedulingGates","value": []}]'

	// patchData := []byte(`[
	// 	{
	// 		"op": "replace",
	// 		"path": "/spec/schedulingGates",
	// 		"value": []
	// 	}
	// ]`)
	// fmt.Println("Removing scheduling gates of first pod")
	// resultPod, err := clientset.CoreV1().Pods(namespace).Patch(context.TODO(), "example-deployment-6876b95d78-crc2q", types.StrategicMergePatchType, patchData, metav1.PatchOptions{},)
	// if err != nil {
	// 	 fmt.Errorf("failed to patch pod %s: %v","pod-name", err)
	// }else {
	// 	fmt.Printf("Pod %s sche gates modified to %s",resultPod.Name,&resultPod.Spec.SchedulingGates)
	// }



	// kubectl patch deployment example-deployment --type=json -p='[{"op": "replace","path": "/spec/template/spec/schedulingGates","value": []}]'
	dPpatchData := []byte(`[
		{
			"op": "replace",
			"path": "/spec/template/spec/schedulingGates",
			"value": []
		}
	]`)

	fmt.Println("Patching deployment")

	_,err=clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), "example-deployment", types.JSONPatchType, dPpatchData, metav1.PatchOptions{})

	if err!=nil{
		fmt.Printf("Error Message: %s\n",err.Error())
	}

	time.Sleep(100*time.Second)

	defer clientset.AppsV1().Deployments(namespace).Delete(context.TODO(),"example-deployment",metav1.DeleteOptions{})


}

func printPod(pod *corev1.Pod){
	fmt.Printf("Pod Name: %s, Status: %s, Phase: %s, Gates: %s\n",pod.Name,&pod.Status.,pod.Status.Phase,pod.Spec.SchedulingGates)
}

func int32Ptr(i int32) *int32 {
	return &i
}