# Kubewarden

## 🅰️ Setup

```
🟢 Start a playground  (using k3s-bare is enough)
labctl playground start k3s-bare
labctl ssh {playground-id}

🟢 clone project
git clone https://github.com/jeffhuang4704/nv2kw.git

🟢 setup via script
~/nv2kw/scripts$ ./setup_kubewarden.sh

```

Detail steps:

```

1️⃣ create namespace kubewarden

kubectl create ns kubewarden

2️⃣ setup helm

helm repo add kubewarden https://charts.kubewarden.io
helm repo update kubewarden

3️⃣ install

helm install --wait -n kubewarden --create-namespace kubewarden-crds kubewarden/kubewarden-crds
helm install --wait -n kubewarden kubewarden-controller kubewarden/kubewarden-controller
helm install --wait -n kubewarden kubewarden-defaults kubewarden/kubewarden-defaults

4️⃣ kw resources

kubectl get all -n kubewarden

admissionpolicies               ap
admissionpolicygroups           apg
clusteradmissionpolicies        cap
clusteradmissionpolicygroups    capg
policyservers                   ps

```

## 🅱️ Test

[Scripts can be found at](../scripts/kw/2test/)

1️⃣ define a ClusterAdmissionPolicy

```
kubectl apply -f - <<EOF
apiVersion: policies.kubewarden.io/v1
kind: ClusterAdmissionPolicy
metadata:
  name: privileged-pods
spec:
  module: registry://ghcr.io/kubewarden/policies/pod-privileged:v0.2.2
  rules:
  - apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
    operations:
    - CREATE
    - UPDATE
  mutating: false
EOF
```

2️⃣✔️ create a non-privilege pod (this is expected to be created okay)

```
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: unprivileged-pod
spec:
  containers:
    - name: nginx
      image: nginx:latest
EOF
```

👉 need to wait when the policy is in active state

```
kubectl get cap --watch
```

3️⃣ ❌ create a privilege pod (this is expected to be denied)

```
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: privileged-pod
spec:
  containers:
    - name: nginx
      image: nginx:latest
      securityContext:
          privileged: true
EOF

```

## Uninstall

```

helm uninstall --namespace kubewarden kubewarden-defaults
helm uninstall --namespace kubewarden kubewarden-controller
helm uninstall --namespace kubewarden kubewarden-crds
kubectl delete namespace kubewarden

```

