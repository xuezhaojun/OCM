IMAGE_TAG=e2e make images

# prepare for hub1
kind create cluster  --name=hub1

kind load docker-image quay.io/open-cluster-management/registration-operator:e2e --name=hub1
kind load docker-image quay.io/open-cluster-management/registration:e2e --name=hub1
kind load docker-image quay.io/open-cluster-management/placement:e2e --name=hub1
kind load docker-image quay.io/open-cluster-management/work:e2e --name=hub1
kind load docker-image quay.io/open-cluster-management/addon-manager:e2e --name=hub1

kind get kubeconfig --name=hub1 > ./hub1
kind get kubeconfig --internal --name=hub1 > ./hub1-kubeconfig
make deploy-hub-operator apply-hub-cr KUBECONFIG=./hub1 IMAGE_TAG=e2e

# prepare for hub2
kind create cluster --name=hub2

kind load docker-image quay.io/open-cluster-management/registration-operator:e2e --name=hub2
kind load docker-image quay.io/open-cluster-management/registration:e2e --name=hub2
kind load docker-image quay.io/open-cluster-management/placement:e2e --name=hub2
kind load docker-image quay.io/open-cluster-management/work:e2e --name=hub2
kind load docker-image quay.io/open-cluster-management/addon-manager:e2e --name=hub2

kind get kubeconfig --name=hub2 > ./hub2
kind get kubeconfig --internal --name=hub2 > ./hub2-kubeconfig
make deploy-hub-operator apply-hub-cr KUBECONFIG=./hub2 IMAGE_TAG=e2e

# prepare for spoke
kind create cluster --name=spoke

kind load docker-image quay.io/open-cluster-management/registration-operator:e2e --name=spoke
kind load docker-image quay.io/open-cluster-management/registration:e2e --name=spoke
kind load docker-image quay.io/open-cluster-management/placement:e2e --name=spoke
kind load docker-image quay.io/open-cluster-management/work:e2e --name=spoke
kind load docker-image quay.io/open-cluster-management/addon-manager:e2e --name=spoke

kind get kubeconfig --name=spoke > ./spoke
make deploy-spoke-operator KUBECONFIG=./spoke IMAGE_TAG=e2e

# deploy multiple bootstrapkubeconfig secrets on the spoke cluster
# bootstrap-secret command install the secret constantly to the namespace "open-cluster-management-agent"
make bootstrap-secret HUB_KUBECONFIG=./hub1-kubeconfig HUB_KUBECONFIG_SECRET_NAME=bootstraphub1 IMAGE_TAG=e2e
make bootstrap-secret HUB_KUBECONFIG=./hub2-kubeconfig HUB_KUBECONFIG_SECRET_NAME=bootstraphub2 IMAGE_TAG=e2e

# run e2e
make run-e2e-multiplehubs \
    HUB1_KUBECONFIG=./hub1 \
    HUB2_KUBECONFIG=./hub2 \
    SPOKE_KUBECONFIG=./spoke \
    IMAGE_TAG=e2e
