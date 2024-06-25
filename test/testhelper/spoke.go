package testhelper

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SpokeTestHelper struct {
	*OCMClients
	klusterletOperatorNamespace string
	klusterletOperator          string
}

func NewSpokeTestHelper(clients *OCMClients) *SpokeTestHelper {
	return &SpokeTestHelper{
		OCMClients: clients,
		// the name of the KlusterletOperator object is constantly "klusterlet-operator" at the moment;
		// The same name as deploy/klusterlet/config/operator/operator.yaml
		klusterletOperatorNamespace: "open-cluster-management",
		klusterletOperator:          "klusterlet",
	}
}

func (h *SpokeTestHelper) CheckKlusterletOperatorReady() error {
	// make sure klusterlet operator deployment is created
	_, err := h.KubeClient.AppsV1().Deployments(h.klusterletOperatorNamespace).
		Get(context.TODO(), h.klusterletOperator, metav1.GetOptions{})
	return err
}
