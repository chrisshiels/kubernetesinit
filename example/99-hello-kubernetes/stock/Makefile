all:
	git clone https://github.com/paulbouwer/hello-kubernetes.git
	( cd hello-kubernetes ; git checkout v1.10.0 )
	helm package hello-kubernetes/deploy/helm/hello-kubernetes/
	helm -n hello-kubernetes template hello-kubernetes \
		./hello-kubernetes-1.0.0.tgz \
		--set deployment.replicaCount=3 \
		--set message=Hello \
		--set service.type=ClusterIP \
		> template.yaml
	rm -rf hello-kubernetes
	rm -f hello-kubernetes-1.0.0.tgz
