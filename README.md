# kubernetesinit

Single-run initialisation of running Kubernetes cluster with CNI, CSI,
Ingress controllers, etc.


## Build

    host$ go build kubernetesinit.go


## Setup

    host$ tree example/
    example/
    ├── 75-metrics-server
    │   ├── all
    │   │   ├── kustomization.yaml
    │   │   └── patch.yaml
    │   ├── kubernetesinit.yaml
    │   ├── local
    │   │   └── kustomization.yaml
    │   └── stock
    │       ├── components.yaml
    │       ├── kustomization.yaml
    │       └── Makefile
    ├── 85-dashboard
    │   ├── all
    │   │   ├── clusterrolebinding.yaml
    │   │   ├── kustomization.yaml
    │   │   └── patch.yaml
    │   ├── kubernetesinit.yaml
    │   ├── local
    │   │   ├── ingress.yaml
    │   │   └── kustomization.yaml
    │   └── stock
    │       ├── kustomization.yaml
    │       ├── Makefile
    │       └── recommended.yaml
    └── 99-hello-kubernetes
        ├── all
        │   ├── kustomization.yaml
        │   └── namespace.yaml
        ├── kubernetesinit.yaml
        ├── local
        │   ├── ingress.yaml
        │   └── kustomization.yaml
        └── stock
            ├── kustomization.yaml
            ├── Makefile
            └── template.yaml

    12 directories, 24 files


## Dry run

    host$ ./kubernetesinit -directory ./example/ -environment local --dryrun
    Run:  kustomize build example/75-metrics-server/local
    Run:  kubectl apply -f -
    Run:  kubectl -n kube-system rollout status deployment metrics-server


    Run:  kustomize build example/85-dashboard/local
    Run:  kubectl apply -f -
    Run:  kubectl -n kubernetes-dashboard rollout status deployment dashboard-metrics-scraper
    Run:  kubectl -n kubernetes-dashboard rollout status deployment kubernetes-dashboard


    Run:  kustomize build example/99-hello-kubernetes/local
    Run:  kubectl apply -f -
    Run:  kubectl -n hello-kubernetes rollout status deployment/hello-kubernetes-hello-kubernetes


## Run

    host$ ./kubernetesinit -directory ./example/ -environment local
    Run:  kustomize build example/75-metrics-server/local
    Run:  kubectl apply -f -
    serviceaccount/metrics-server created
    clusterrole.rbac.authorization.k8s.io/system:aggregated-metrics-reader created
    clusterrole.rbac.authorization.k8s.io/system:metrics-server created
    rolebinding.rbac.authorization.k8s.io/metrics-server-auth-reader created
    clusterrolebinding.rbac.authorization.k8s.io/metrics-server:system:auth-delegator created
    clusterrolebinding.rbac.authorization.k8s.io/system:metrics-server created
    service/metrics-server created
    deployment.apps/metrics-server created
    apiservice.apiregistration.k8s.io/v1beta1.metrics.k8s.io created
    Run:  kubectl -n kube-system rollout status deployment metrics-server
    Waiting for deployment "metrics-server" rollout to finish: 0 of 1 updated replicas are available...
    deployment "metrics-server" successfully rolled out


    Run:  kustomize build example/85-dashboard/local
    Run:  kubectl apply -f -
    namespace/kubernetes-dashboard created
    serviceaccount/kubernetes-dashboard created
    role.rbac.authorization.k8s.io/kubernetes-dashboard created
    clusterrole.rbac.authorization.k8s.io/kubernetes-dashboard created
    rolebinding.rbac.authorization.k8s.io/kubernetes-dashboard created
    clusterrolebinding.rbac.authorization.k8s.io/kubernetes-dashboard-cluster-admin created
    clusterrolebinding.rbac.authorization.k8s.io/kubernetes-dashboard created
    configmap/kubernetes-dashboard-settings created
    secret/kubernetes-dashboard-certs created
    secret/kubernetes-dashboard-csrf created
    secret/kubernetes-dashboard-key-holder created
    service/dashboard-metrics-scraper created
    service/kubernetes-dashboard created
    deployment.apps/dashboard-metrics-scraper created
    deployment.apps/kubernetes-dashboard created
    ingress.networking.k8s.io/kubernetes-dashboard created
    Run:  kubectl -n kubernetes-dashboard rollout status deployment dashboard-metrics-scraper
    Waiting for deployment "dashboard-metrics-scraper" rollout to finish: 0 of 1 updated replicas are available...
    deployment "dashboard-metrics-scraper" successfully rolled out
    Run:  kubectl -n kubernetes-dashboard rollout status deployment kubernetes-dashboard
    Waiting for deployment "kubernetes-dashboard" rollout to finish: 0 of 1 updated replicas are available...
    deployment "kubernetes-dashboard" successfully rolled out


    Run:  kustomize build example/99-hello-kubernetes/local
    Run:  kubectl apply -f -
    namespace/hello-kubernetes created
    serviceaccount/hello-kubernetes-hello-kubernetes created
    service/hello-kubernetes-hello-kubernetes created
    deployment.apps/hello-kubernetes-hello-kubernetes created
    ingress.networking.k8s.io/hello-kubernetes created
    Run:  kubectl -n hello-kubernetes rollout status deployment/hello-kubernetes-hello-kubernetes
    Waiting for deployment "hello-kubernetes-hello-kubernetes" rollout to finish: 0 of 3 updated replicas are available...
    Waiting for deployment "hello-kubernetes-hello-kubernetes" rollout to finish: 1 of 3 updated replicas are available...
    Waiting for deployment "hello-kubernetes-hello-kubernetes" rollout to finish: 2 of 3 updated replicas are available...
    deployment "hello-kubernetes-hello-kubernetes" successfully rolled out
