kubectl create ns kubewarden

helm repo add kubewarden https://charts.kubewarden.io
helm repo update kubewarden

helm install --wait -n kubewarden --create-namespace kubewarden-crds kubewarden/kubewarden-crds
helm install --wait -n kubewarden kubewarden-controller kubewarden/kubewarden-controller
helm install --wait -n kubewarden kubewarden-defaults kubewarden/kubewarden-defaults

kubectl get pods -n kubewarden --watch

