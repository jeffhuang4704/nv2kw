echo '🔴  Be aware of the annotation behavior !'
echo '🔴 🔴 k8s will add annotation automatically. So in operator=container_other_than will probably triggered all the time'
kubectl create -f 1_policy1.yaml
kubectl create -f 2_policy1_binding.yaml
