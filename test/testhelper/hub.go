package testhelper

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocmfeature "open-cluster-management.io/api/feature"
	operatorapiv1 "open-cluster-management.io/api/operator/v1"

	"open-cluster-management.io/ocm/pkg/operator/helpers"
)

type HubTestHelper struct {
	*OCMClients
	clusterManagerName      string
	clusterManagerNamespace string
}

func NewHubTestHelper(clients *OCMClients) *HubTestHelper {
	return &HubTestHelper{
		OCMClients: clients,
		// the name of the ClusterManager object is constantly "cluster-manager" at the moment; The same name as deploy/cluster-manager/config/samples
		clusterManagerName:      "cluster-manager",
		clusterManagerNamespace: helpers.ClusterManagerDefaultNamespace,
	}
}

func (h *HubTestHelper) EnableAutoApprove(users []string) error {
	cm, err := h.getCluserManager(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get cluster manager: %w", err)
	}
	if cm.Spec.RegistrationConfiguration == nil {
		cm.Spec.RegistrationConfiguration = &operatorapiv1.RegistrationHubConfiguration{}
	}
	cm.Spec.RegistrationConfiguration.FeatureGates = append(cm.Spec.RegistrationConfiguration.FeatureGates, operatorapiv1.FeatureGate{
		Feature: string(ocmfeature.ManagedClusterAutoApproval),
		Mode:    operatorapiv1.FeatureGateModeTypeEnable,
	})
	cm.Spec.RegistrationConfiguration.AutoApproveUsers = users
	_, err = h.OperatorClient.OperatorV1().ClusterManagers().Update(context.TODO(), cm, metav1.UpdateOptions{})
	return err
}

func (h *HubTestHelper) CheckHubReady(ctx context.Context) error {
	cm, err := h.getCluserManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster manager: %w", err)
	}

	err = checkClusterManagerStatus(cm)
	if err != nil {
		return fmt.Errorf("failed to check cluster manager status: %w", err)
	}

	// make sure open-cluster-management-hub namespace is created
	if _, err := h.KubeClient.CoreV1().Namespaces().
		Get(context.TODO(), h.clusterManagerNamespace, metav1.GetOptions{}); err != nil {
		return err
	}

	// make sure deployments are ready
	deployments := []string{
		fmt.Sprintf("%s-registration-controller", h.clusterManagerName),
		fmt.Sprintf("%s-registration-webhook", h.clusterManagerName),
		fmt.Sprintf("%s-work-webhook", h.clusterManagerName),
		fmt.Sprintf("%s-placement-controller", h.clusterManagerName),
	}
	for _, deployment := range deployments {
		if err = CheckDeploymentReady(ctx, h.KubeClient, h.clusterManagerNamespace, deployment); err != nil {
			return fmt.Errorf("failed to check deployment %s: %w", deployment, err)
		}
	}

	// if manifestworkreplicaset feature is enabled, check the work controller
	if cm.Spec.WorkConfiguration != nil &&
		helpers.FeatureGateEnabled(cm.Spec.WorkConfiguration.FeatureGates, ocmfeature.DefaultHubWorkFeatureGates, ocmfeature.ManifestWorkReplicaSet) {
		if err = CheckDeploymentReady(ctx, h.KubeClient, h.clusterManagerNamespace, fmt.Sprintf("%s-work-controller", h.clusterManagerName)); err != nil {
			return fmt.Errorf("failed to check work controller: %w", err)
		}
	}

	// if addonManager feature is enabled, check the addonManager controller
	if cm.Spec.AddOnManagerConfiguration != nil &&
		helpers.FeatureGateEnabled(cm.Spec.AddOnManagerConfiguration.FeatureGates, ocmfeature.DefaultHubAddonManagerFeatureGates, ocmfeature.AddonManagement) {
		if err = CheckDeploymentReady(ctx, h.KubeClient, h.clusterManagerNamespace, fmt.Sprintf("%s-addon-manager-controller", h.clusterManagerName)); err != nil {
			return fmt.Errorf("failed to check addon manager controller: %w", err)
		}
	}

	return nil
}

func (h *HubTestHelper) SetHubAcceptsClient(ctx context.Context, clusterName string, hubAcceptClient bool) error {
	managedCluster, err := h.ClusterClient.ClusterV1().ManagedClusters().Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get managed cluster: %w", err)
	}

	if managedCluster.Spec.HubAcceptsClient != hubAcceptClient {
		managedCluster.Spec.HubAcceptsClient = hubAcceptClient
		_, err = h.ClusterClient.ClusterV1().ManagedClusters().Update(ctx, managedCluster, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update managed cluster: %w", err)
		}
	}

	return nil
}

func (h *HubTestHelper) SetLeaseDurationSeconds(ctx context.Context, clusterName string, leaseDurationSeconds int32) error {
	managedCluster, err := h.ClusterClient.ClusterV1().ManagedClusters().Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get managed cluster: %w", err)
	}

	managedCluster.Spec.LeaseDurationSeconds = leaseDurationSeconds
	_, err = h.ClusterClient.ClusterV1().ManagedClusters().Update(ctx, managedCluster, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update managed cluster: %w", err)
	}
	return nil
}

// getCluserManager returns the ClusterManager object from the hub
func (h *HubTestHelper) getCluserManager(ctx context.Context) (*operatorapiv1.ClusterManager, error) {
	return h.OperatorClient.OperatorV1().ClusterManagers().Get(ctx, h.clusterManagerName, metav1.GetOptions{})
}

func checkClusterManagerStatus(cm *operatorapiv1.ClusterManager) error {
	if meta.IsStatusConditionFalse(cm.Status.Conditions, "Applied") {
		return fmt.Errorf("components of cluster manager are not all applied")
	}
	if meta.IsStatusConditionFalse(cm.Status.Conditions, "ValidFeatureGates") {
		return fmt.Errorf("feature gates are not all valid")
	}
	if !meta.IsStatusConditionFalse(cm.Status.Conditions, "HubRegistrationDegraded") {
		return fmt.Errorf("HubRegistration is degraded")
	}
	if !meta.IsStatusConditionFalse(cm.Status.Conditions, "HubPlacementDegraded") {
		return fmt.Errorf("HubPlacement is degraded")
	}
	if !meta.IsStatusConditionFalse(cm.Status.Conditions, "Progressing") {
		return fmt.Errorf("ClusterManager is still progressing")
	}
	return nil
}
