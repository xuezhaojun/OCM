package testhelper

import (
	"context"
	"fmt"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	addonclient "open-cluster-management.io/api/client/addon/clientset/versioned"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	operatorclient "open-cluster-management.io/api/client/operator/clientset/versioned"
	workv1client "open-cluster-management.io/api/client/work/clientset/versioned"
)

type Images struct {
	RegistrationImage string
	WorkImage         string
	SingletonImage    string
}

// OCMClients contains every kind of client that we need to interact with ocm components
type OCMClients struct {
	KubeClient          kubernetes.Interface
	APIExtensionsClient apiextensionsclient.Interface
	OperatorClient      operatorclient.Interface
	ClusterClient       clusterclient.Interface
	WorkClient          workv1client.Interface
	AddonClient         addonclient.Interface
	DynamicClient       dynamic.Interface
	RestMapper          meta.RESTMapper
}

func NewOCMClients(kubeConfigPath string) (*OCMClients, error) {
	clusterCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build managed cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster client: %w", err)
	}

	httpClient, err := rest.HTTPClientFor(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster http client: %w", err)
	}

	restMapper, err := apiutil.NewDynamicRESTMapper(clusterCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster rest mapper: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster dynamic client: %w", err)
	}

	apiExtensionsClient, err := apiextensionsclient.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster api extensions client: %w", err)
	}

	operatorClient, err := operatorclient.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster operator client: %w", err)
	}

	clusterClient, err := clusterclient.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster cluster client: %w", err)
	}

	workClient, err := workv1client.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster work client: %w", err)
	}

	addonClient, err := addonclient.NewForConfig(clusterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed cluster addon client: %w", err)
	}

	return &OCMClients{
		KubeClient:          kubeClient,
		APIExtensionsClient: apiExtensionsClient,
		OperatorClient:      operatorClient,
		ClusterClient:       clusterClient,
		WorkClient:          workClient,
		AddonClient:         addonClient,
		DynamicClient:       dynamicClient,
		RestMapper:          restMapper,
	}, nil
}

func CheckDeploymentReady(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string) error {
	deployment, err := kubeClient.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", name, err)
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return fmt.Errorf("deployment %s is not ready, ready replicas: %d, replicas: %d", name, deployment.Status.ReadyReplicas, deployment.Status.Replicas)
	}

	return nil
}

func (h *HubTestHelper) CheckManagedClusterStatusConditions(ctx context.Context, clusterName string,
	expectedConditions map[string]metav1.ConditionStatus) error {
	if clusterName == "" {
		return fmt.Errorf("the name of managedcluster should not be null")
	}

	cluster, err := h.ClusterClient.ClusterV1().ManagedClusters().Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// expect the managed cluster to be not available
	for conditionType, conditionStatus := range expectedConditions {
		condition := meta.FindStatusCondition(cluster.Status.Conditions, conditionType)
		if condition == nil {
			return fmt.Errorf("managed cluster %s is not in expected status, expect %s to be %s, but not found",
				clusterName, conditionType, conditionStatus)
		}
		if condition.Status != conditionStatus {
			return fmt.Errorf("managed cluster %s is not in expected status, expect %s to be %s, but got %s",
				clusterName, conditionType, conditionStatus, condition.Status)
		}
	}

	return nil
}
