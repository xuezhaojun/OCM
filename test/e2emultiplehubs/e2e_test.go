package e2emultiplehubs

import (
	"context"
	"flag"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"open-cluster-management.io/ocm/test/testhelper"
)

const (
	// The secrets' names are the same as the secrets created in GitHub Actions: e2e-multiplehubs
	HUB1_BOOTSTRAP_SECRET_NAME = "bootstraphub1" //nolint:gosec
	HUB2_BOOTSRTAP_SECRET_NAME = "bootstraphub2" //nolint:gosec
	clusterName                = "spoke"
)

var (
	registrationImage string
	workImage         string
	singletonImage    string

	hub1Kubeconfig  string
	hub2Kubeconfig  string
	spokeKubeconfig string
)

var (
	hub1TestHelper  *testhelper.HubTestHelper
	hub2TestHelper  *testhelper.HubTestHelper
	spokeTestHelper *testhelper.SpokeTestHelper
)

func init() {
	flag.StringVar(&registrationImage, "registration-image", "", "The image of the registration")
	flag.StringVar(&workImage, "work-image", "", "The image of the work")
	flag.StringVar(&singletonImage, "singleton-image", "", "The image of the klusterlet agent")

	flag.StringVar(&hub1Kubeconfig, "hub1-kubeconfig", "", "The kubeconfig of the hub cluster 1")
	flag.StringVar(&hub2Kubeconfig, "hub2-kubeconfig", "", "The kubeconfig of the hub cluster 2")
	flag.StringVar(&spokeKubeconfig, "spoke-kubeconfig", "", "The kubeconfig of the spoke cluster")
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MuE2E Suite")
}

var _ = BeforeSuite(func() {
	var err error

	// setup hubTestHelper for hub1
	By("setup hubTestHelper for hub1")
	hub1Clients, err := testhelper.NewOCMClients(hub1Kubeconfig)
	Expect(err).ToNot(HaveOccurred())

	hub1TestHelper = testhelper.NewHubTestHelper(hub1Clients)

	// setup hubTestHelper for hub2
	By("setup hubTestHelper for hub2")
	hub2Clients, err := testhelper.NewOCMClients(hub2Kubeconfig)
	Expect(err).ToNot(HaveOccurred())

	hub2TestHelper = testhelper.NewHubTestHelper(hub2Clients)

	// setup agentTestHelper for spoke
	By("setup agentTestHelper for spoke")
	spokeClients, err := testhelper.NewOCMClients(spokeKubeconfig)
	Expect(err).ToNot(HaveOccurred())

	spokeTestHelper = testhelper.NewSpokeTestHelper(spokeClients)

	// check hub1 ready
	By("check hub1 ready")
	Eventually(func() error {
		return hub1TestHelper.CheckHubReady(context.TODO())
	}, 5*time.Minute, 5*time.Second).Should(Succeed())

	// check hub2 ready
	By("check hub2 ready")
	Eventually(func() error {
		return hub2TestHelper.CheckHubReady(context.TODO())
	}, 5*time.Minute, 5*time.Second).Should(Succeed())

	// check klusterlet operator ready
	By("check klusterlet operator ready")
	Eventually(spokeTestHelper.CheckKlusterletOperatorReady, 5*time.Minute, 5*time.Second).Should(Succeed())

	// enable AutoApprove on hub1
	By("enable AutoApprove on hub1")
	Eventually(func() error {
		return hub1TestHelper.EnableAutoApprove([]string{"kubernetes-admin"})
	}, 5*time.Minute, 5*time.Second).Should(Succeed())

	// enable AutoApprove on hub2
	By("enable AutoApprove on hub2")
	Eventually(func() error {
		return hub2TestHelper.EnableAutoApprove([]string{"kubernetes-admin"})
	}, 5*time.Minute, 5*time.Second).Should(Succeed())
})
