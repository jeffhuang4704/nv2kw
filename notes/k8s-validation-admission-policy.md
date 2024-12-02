# k8s Validating Admission Policy

## resources
- [k8s doc](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)

## how it works

(quote from k8s doc)

*What Resources Make a Policy*  

A policy is generally made up of three resources:

- The `ValidatingAdmissionPolicy` describes the abstract logic of a policy (think: "this policy makes sure a particular label is set to a particular value").

- A `ValidatingAdmissionPolicyBinding` links the above resources together and provides scoping. If you only want to require an owner label to be set for Pods, the binding is where you would specify this restriction.

- A parameter resource provides information to a `ValidatingAdmissionPolicy` to make it a concrete statement (think "the owner label must be set to something that ends in .company.com"). A native type such as ConfigMap or a CRD defines the schema of a parameter resource. `ValidatingAdmissionPolicy` objects specify what Kind they are expecting for their parameter resource.

At least a `ValidatingAdmissionPolicy` and a corresponding `ValidatingAdmissionPolicyBinding` must be defined for a policy to have an effect.

If a `ValidatingAdmissionPolicy` does not need to be configured via parameters, simply leave spec.paramKind in `ValidatingAdmissionPolicy` not specified.


## start minikube with feature enabled

```
# start minikube with the enable flag
minikube start --feature-gates=ValidatingAdmissionPolicy=true

# verify
neuvector@ubuntu2204-E:~$ kubectl logs -n kube-system $(kubectl get pods -n kube-system -l component=kube-apiserver -o name) | grep "ValidatingAdmissionPolicy"
W1201 19:20:02.179525       1 feature_gate.go:246] Setting GA feature gate ValidatingAdmissionPolicy=true. It will be removed in a future release.
I1201 19:20:03.672822       1 shared_informer.go:313] Waiting for caches to sync for *generic.policySource[*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicy,*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicyBinding,k8s.io/apiserver/pkg/admission/plugin/policy/validating.Validator]
I1201 19:20:03.675500       1 plugins.go:160] Loaded 13 validating admission controller(s) successfully in the following order: LimitRanger,ServiceAccount,PodSecurity,Priority,PersistentVolumeClaimResize,RuntimeClass,CertificateApproval,CertificateSigning,ClusterTrustBundleAttest,CertificateSubjectRestriction,ValidatingAdmissionPolicy,ValidatingAdmissionWebhook,ResourceQuota.
I1201 19:20:08.019776       1 shared_informer.go:320] Caches are synced for *generic.policySource[*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicy,*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicyBinding,k8s.io/apiserver/pkg/admission/plugin/policy/validating.Validator]
```

## testing 1 - single expression

We need `Policy` and `Binding` resources - `ValidatingAdmissionPolicy` and `ValidatingAdmissionPolicyBinding`

*Policy*

<details><summary>YAML</summary>

```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "demo-policy.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["apps"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["deployments"]
  validations:
    - expression: "object.spec.replicas <= 5"
```
</details>

*Binding*

<details><summary>YAML</summary>

```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: "demo-binding-test.example.com"
spec:
  policyName: "demo-policy.example.com"
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        environment: test     👈

```
</details>

<details><summary>Get resource</summary>

```
neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl get ValidatingAdmissionPolicy
NAME                      VALIDATIONS   PARAMKIND   AGE
demo-policy.example.com   1             <unset>     111m

neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl get ValidatingAdmissionPolicyBinding
NAME                            POLICYNAME                PARAMREF   AGE
demo-binding-test.example.com   demo-policy.example.com   <unset>    111m

neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl get ValidatingAdmissionPolicy demo-policy.example.com -oyaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: demo-policy.example.com
  resourceVersion: "603"
  uid: aaf05774-6191-4fca-a78c-aca19f5d981e
spec:
  failurePolicy: Fail
  matchConstraints:
    matchPolicy: Equivalent
    namespaceSelector: {}
    objectSelector: {}
    resourceRules:
    - apiGroups:
      - apps
      apiVersions:
      - v1
      operations:
      - CREATE
      - UPDATE
      resources:
      - deployments
      scope: '*'
  validations:
  - expression: object.spec.replicas <= 5
status:
  observedGeneration: 1
  typeChecking: {}


neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl get ValidatingAdmissionPolicyBinding demo-binding-test.example.com -oyaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: demo-binding-test.example.com
spec:
  matchResources:
    matchPolicy: Equivalent
    namespaceSelector:
      matchLabels:
        environment: test
    objectSelector: {}
  policyName: demo-policy.example.com
  validationActions:
  - Deny
```
</details>

🔴🔴🔴 Make sure your testing environment has the matched criteria (matched label `environment=test`)

The `namespaceSelector` in your `ValidatingAdmissionPolicyBinding` requires namespaces with the label `environment: test`.
If the test namespace does not have this label, the policy won't apply to resources in the test namespace.

*Add label the namespace*

```
# add label to the namespace
kubectl label namespace test environment=test

# get the namespace 
neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl get ns test -oyaml
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: "2024-12-01T18:38:40Z"
  labels:
    environment: test  👈
    kubernetes.io/metadata.name: test
  name: test

```

*Resource*

<details><summary>YAML</summary>

```
neuvector@ubuntu2204-E:~/validating_admission_policy$ cat deploy1.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
  name: my-dep
  namespace: test     👈
spec:
  replicas: 6
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: nginx
        name: nginx
        resources: {}
status: {}
```
</details>

*Deployment denied*
```
neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl apply -f deploy1.yaml
The deployments "my-dep" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: failed expression: object.spec.replicas <= 5
```

## testing 2 - multiple expressions

<details><summary>YAML</summary>

```
# policy
neuvector@ubuntu2204-E:~/validating_admission_policy/2_multi_expressions$ cat policy2.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "demo-policy.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["apps"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["deployments"]
  validations:
    - expression: "object.spec.replicas <= 5"                    🔴
      message: "The number of replicas must not exceed 5."
    - expression: "object.spec.template.spec.containers.all(c, !c.image.endsWith(':latest'))"  🔴
      message: "Container images must not use the 'latest' tag."

# binding
neuvector@ubuntu2204-E:~/validating_admission_policy/2_multi_expressions$ cat policy2_binding.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: "demo-binding-test.example.com"
spec:
  policyName: "demo-policy.example.com"
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        environment: test

# resource
neuvector@ubuntu2204-E:~/validating_admission_policy/2_multi_expressions$ cat deploy2.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prod-deployment
  labels:
    env: prod
  namespace: test
spec:
  replicas: 4
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        resources:
          requests:
            cpu: "0.5"  # CPU requests are not in millicores

# deployment denied 1 (in this case the replicas=4)
neuvector@ubuntu2204-E:~/validating_admission_policy/2_multi_expressions$ kubectl apply -f deploy2.yaml
The deployments "prod-deployment" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: Container images must not use the 'latest' tag.

# deployment denied 2 (in this case the replicas=10)
neuvector@ubuntu2204-E:~/validating_admission_policy/2_multi_expressions$ kubectl apply -f deploy2.yaml
The deployments "prod-deployment" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: The number of replicas must not exceed 5.
```

</details>

## testing 3 - multiple lines in a single expression

<details><summary>YAML</summary>

```
# policy
neuvector@ubuntu2204-E:~/validating_admission_policy/3_multi_lines_expression$ cat policy3.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "demo-policy.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["apps"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["deployments"]
  validations:
    - expression: |   🔴
        object.spec.replicas <= 5 &&
        object.spec.template.spec.containers.all(c,
          c.image.startsWith('myregistry.io/') &&
          !c.image.endsWith(':latest')
        )
      message: >      🔴
        The number of replicas must not exceed 5.
        All container images must be from 'myregistry.io' and must not use the 'latest' tag.


# binding        
neuvector@ubuntu2204-E:~/validating_admission_policy/3_multi_lines_expression$ cat policy3_binding.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: "demo-binding-test.example.com"
spec:
  policyName: "demo-policy.example.com"
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        environment: test

# resource
neuvector@ubuntu2204-E:~/validating_admission_policy/3_multi_lines_expression$ cat deploy3.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prod-deployment
  labels:
    env: prod
  namespace: test
spec:
  replicas: 10
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        resources:
          requests:
            cpu: "0.5"  # CPU requests are not in millicores

# deployment denied
neuvector@ubuntu2204-E:~/validating_admission_policy/3_multi_lines_expression$ kubectl apply -f deploy3.yaml
The deployments "prod-deployment" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: The number of replicas must not exceed 5. All container images must be from 'myregistry.io' and must not use the 'latest' tag.
```

</details>

## testing 4 - resource limit check

🔴 TODO: Need to consider the Memory resource units, see the [doc](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
Also check [Kubewarden's Container resources policy](https://github.com/kubewarden/container-resources-policy)

<details><summary>YAML</summary>

```
# policy
neuvector@ubuntu2204-E:~/validating_admission_policy/5_resource_limit$ cat policy5.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "demo-policy.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["apps"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["deployments"]
  validations:
    - expression: >
        ['Deployment','ReplicaSet','DaemonSet','StatefulSet','Job'].all(kind, object.kind != kind) ||
        object.spec.template.spec.containers.all(container,
          (has(container.resources) &&
           has(container.resources.requests) &&
           has(container.resources.requests.memory) &&
           100 * 1024 * 1024 <= int(container.resources.requests.memory.replace("Mi", "").replace("Gi", "000")) * 1024 * 1024 &&
           200 * 1024 * 1024 >= int(container.resources.requests.memory.replace("Mi", "").replace("Gi", "000")) * 1024 * 1024) &&
          (has(container.resources.limits) &&
           has(container.resources.limits.memory) &&
           100 * 1024 * 1024 <= int(container.resources.limits.memory.replace("Mi", "").replace("Gi", "000")) * 1024 * 1024 &&
           200 * 1024 * 1024 >= int(container.resources.limits.memory.replace("Mi", "").replace("Gi", "000")) * 1024 * 1024)
        )
      message: "Workloads contain containers with memory limits or requests not set, or they are not in the specified range (100Mi to 200Mi)."


# binding
neuvector@ubuntu2204-E:~/validating_admission_policy/5_resource_limit$ cat policy5_binding.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: "demo-binding-test.example.com"
spec:
  policyName: "demo-policy.example.com"
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        environment: test


# deployment denied
neuvector@ubuntu2204-E:~/validating_admission_policy/5_resource_limit$ kubectl apply -f deploy5.yaml
The deployments "prod-deployment" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: Workloads contain containers with memory limits or requests not set, or they are not in the specified range (100Mi to 200Mi).
```

</details>
