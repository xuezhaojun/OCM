package e2emultiplehubs

import (
	"context"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ocmfeature "open-cluster-management.io/api/feature"
	operatorapiv1 "open-cluster-management.io/api/operator/v1"

	"open-cluster-management.io/ocm/pkg/operator/helpers"
)

// The test cases and steps are strictly ordered:
// #1: the spoke cluster registry to the hub1
// #2: set `hubAcceptClient` to false on the hub1, expect to see spoke switch to registry to the hub2
// #3: set `hubAcceptClient` to true on the hub1, remove the managedcluser spoke on hub1, then shutdown
// the hub2, expect to see spoke switch to registry to the hub1
// Notes:
// * The namesapce "open-cluster-management-agent" is already created in the e2e.yml, the "make bootstrap-secret" step.
var _ = Describe("MultipleHubs Test", Ordered, func() {
	It("Spoke cluster registries to the hub1", func() {
		var err error

		// configure klusterlet
		klusterlet := operatorapiv1.Klusterlet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "klusterlet",
			},
			Spec: operatorapiv1.KlusterletSpec{
				ClusterName: clusterName,
				Namespace:   helpers.KlusterletDefaultNamespace,
				// images
				RegistrationImagePullSpec: registrationImage,
				WorkImagePullSpec:         workImage,
				ImagePullSpec:             singletonImage,
				// using singleton mode to save resources during the test
				DeployOption: operatorapiv1.KlusterletDeployOption{
					Mode: operatorapiv1.InstallModeSingleton,
				},
				RegistrationConfiguration: &operatorapiv1.RegistrationConfiguration{
					FeatureGates: []operatorapiv1.FeatureGate{
						{
							Feature: string(ocmfeature.MultipleHubs),
							Mode:    operatorapiv1.FeatureGateModeTypeEnable,
						},
					},
					BootstrapKubeConfigs: operatorapiv1.BootstrapKubeConfigs{
						Type: operatorapiv1.LocalSecrets,
						LocalSecrets: operatorapiv1.LocalSecretsConfig{
							KubeConfigSecrets: []operatorapiv1.KubeConfigSecret{
								{
									Name: HUB1_BOOTSTRAP_SECRET_NAME,
								},
								{
									Name: HUB2_BOOTSRTAP_SECRET_NAME,
								},
							},
							HubConnectionTimeoutSeconds: 3 * 60, // this value by default is 10m, change it to 3m to trigger reselect quickly
						},
					},
				},
			},
		}

		// create klusterlet
		_, err = spokeTestHelper.OperatorClient.OperatorV1().Klusterlets().Create(context.TODO(), &klusterlet, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())

		// mangedcluster should be Available on the hub1
		Eventually(func() error {
			return hub1TestHelper.CheckManagedClusterStatusConditions(context.TODO(), clusterName, map[string]metav1.ConditionStatus{
				clusterv1.ManagedClusterConditionAvailable:   metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionJoined:      metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionHubAccepted: metav1.ConditionTrue,
			})
		}, 5*time.Minute, 5*time.Second).Should(Succeed())
	})

	It("Spoke cluster switches to hub2 when <hubAcceptClient> set to false on hub1", func() {
		// update the leaseDuration to 2s to make agent update lease more frequently
		// it also means the hub will check the managed cluster lease every 10s
		Expect(hub1TestHelper.SetLeaseDurationSeconds(context.TODO(), clusterName, 2)).Should(Succeed())

		// set hubAcceptClient to false on hub1
		Expect(hub1TestHelper.SetHubAcceptsClient(context.TODO(), clusterName, false)).Should(Succeed())

		// check the managed cluster to be Available on the hub2
		Eventually(func() error {
			return hub2TestHelper.CheckManagedClusterStatusConditions(context.TODO(), clusterName, map[string]metav1.ConditionStatus{
				clusterv1.ManagedClusterConditionAvailable:   metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionJoined:      metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionHubAccepted: metav1.ConditionTrue,
			})
		}, 5*time.Minute, 5*time.Second).Should(Succeed())

		// TODO: @xuezhaojun after the agent switch to the hub2, the 'managedcluster' on the hub1 keep showing `Available` status
		// This is because:
		// * the lease contoller only watch managedcluster resource, no periodical check on the managedcluster lease
		// * if the `hubAcceptsClient` is false, the controller will skip the reconciliation
		// In the future, we may need to add a periodical check on the managedcluster lease to make sure the managedcluster status is correct
	})

	It("Spoke cluster switches back to hub1 when <hubAcceptClient> set to true on hub1 and hub2 is down", func() {
		var err error

		// delete manaegd cluster from hub1
		Expect(hub1TestHelper.ClusterClient.ClusterV1().ManagedClusters().Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).To(Succeed())

		// eventually get managed cluster not found on hub1
		Eventually(func() bool {
			_, err = hub1TestHelper.ClusterClient.ClusterV1().ManagedClusters().Get(context.TODO(), clusterName, metav1.GetOptions{})
			return apierrors.IsNotFound(err)
		}, 5*time.Minute, 5*time.Second).Should(BeTrue())

		// run "kind delete cluster --name hub2" to shutdown the hub2
		cmd := exec.Command("kind", "delete", "cluster", "--name", "hub2")
		output, err := cmd.Output()
		Expect(err).ToNot(HaveOccurred(), string(output))

		// mangedcluster should be Available on the hub1
		Eventually(func() error {
			return hub1TestHelper.CheckManagedClusterStatusConditions(context.TODO(), clusterName, map[string]metav1.ConditionStatus{
				clusterv1.ManagedClusterConditionAvailable:   metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionJoined:      metav1.ConditionTrue,
				clusterv1.ManagedClusterConditionHubAccepted: metav1.ConditionTrue,
			})
		}, 5*time.Minute, 5*time.Second).Should(Succeed())
	})
})
