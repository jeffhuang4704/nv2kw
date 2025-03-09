we need to create a resource in a namespace that has the label environment=test, since the binding applies only to such namespaces

To add the label environment=test to the ns1 namespace, use the following command:
kubectl label namespace ns1 environment=test
