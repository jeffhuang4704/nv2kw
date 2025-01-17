echo 'cleaning...'
kubectl delete -f 1_policy1.yaml
kubectl delete -f 2_deployment.yaml

echo 'create policy...'
kubectl apply -f 1_policy1.yaml
kubectl get ClusterAdmissionPolicy --watch
