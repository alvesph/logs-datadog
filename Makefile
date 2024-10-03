ADD_OPENTELEMETRY:
	helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

UPDATE_HELM:
	helm repo update

INSTALL_OPENTELEMETRY:
	helm install my-opentelemetry-collector open-telemetry/opentelemetry-collector \
		--set image.repository="otel/opentelemetry-collector-k8s" \
		--set mode=daemonset

INSTALL_COLLECTOR: ADD_OPENTELEMETRY UPDATE_HELM INSTALL_OPENTELEMETRY
	kubectl apply -f ./k9s/collector.yaml
