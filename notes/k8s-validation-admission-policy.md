# k8s Validating Admission Policy

## resources
- [k8s doc](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)

## how it works

*What Resources Make a Policy*
A policy is generally made up of three resources:

- The `ValidatingAdmissionPolicy` describes the abstract logic of a policy (think: "this policy makes sure a particular label is set to a particular value").

- A `ValidatingAdmissionPolicyBinding` links the above resources together and provides scoping. If you only want to require an owner label to be set for Pods, the binding is where you would specify this restriction.

- A parameter resource provides information to a `ValidatingAdmissionPolicy` to make it a concrete statement (think "the owner label must be set to something that ends in .company.com"). A native type such as ConfigMap or a CRD defines the schema of a parameter resource. `ValidatingAdmissionPolicy` objects specify what Kind they are expecting for their parameter resource.

At least a `ValidatingAdmissionPolicy` and a corresponding `ValidatingAdmissionPolicyBinding` must be defined for a policy to have an effect.

If a `ValidatingAdmissionPolicy` does not need to be configured via parameters, simply leave spec.paramKind in `ValidatingAdmissionPolicy` not specified.


## minikube

```
# need to start minikube with the enable flag
minikube start --feature-gates=ValidatingAdmissionPolicy=true

# verify
neuvector@ubuntu2204-E:~$ kubectl logs -n kube-system $(kubectl get pods -n kube-system -l component=kube-apiserver -o name) | grep "ValidatingAdmissionPolicy"
W1201 19:20:02.179525       1 feature_gate.go:246] Setting GA feature gate ValidatingAdmissionPolicy=true. It will be removed in a future release.
I1201 19:20:03.672822       1 shared_informer.go:313] Waiting for caches to sync for *generic.policySource[*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicy,*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicyBinding,k8s.io/apiserver/pkg/admission/plugin/policy/validating.Validator]
I1201 19:20:03.675500       1 plugins.go:160] Loaded 13 validating admission controller(s) successfully in the following order: LimitRanger,ServiceAccount,PodSecurity,Priority,PersistentVolumeClaimResize,RuntimeClass,CertificateApproval,CertificateSigning,ClusterTrustBundleAttest,CertificateSubjectRestriction,ValidatingAdmissionPolicy,ValidatingAdmissionWebhook,ResourceQuota.
I1201 19:20:08.019776       1 shared_informer.go:320] Caches are synced for *generic.policySource[*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicy,*k8s.io/api/admissionregistration/v1.ValidatingAdmissionPolicyBinding,k8s.io/apiserver/pkg/admission/plugin/policy/validating.Validator]
```

## testing

refer to the [k8s doc](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
We need two resources `ValidatingAdmissionPolicy` and `ValidatingAdmissionPolicyBinding`

*Policy *

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

*Binding*

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
        environment: test

```

We need to add the matched lael `environment=test`

The `namespaceSelector` in your `ValidatingAdmissionPolicyBinding` requires namespaces with the label environment: test.
If the test namespace does not have this label, the policy won't apply to resources in the test namespace.

*Add label the namespace: *

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

*Deployment denied*
```
neuvector@ubuntu2204-E:~/validating_admission_policy$ kubectl apply -f deploy1.yaml
The deployments "my-dep" is invalid: : ValidatingAdmissionPolicy 'demo-policy.example.com' with binding 'demo-binding-test.example.com' denied request: failed expression: object.spec.replicas <= 5
```