[ ] docker build -t ws-app ./go
[ ] minikube image load ws-app
[ ] minikube image ls | grep ws
[ ] k apply -f k8s/deployment.yaml
[ ] k apply -f k8s/service.yaml
[ ] helm repo add haproxytech https://haproxytech.github.io/helm-charts
[ ] helm repo update
[ ] docker pull haproxytech/kubernetes-ingress:3.1.14
[ ] docker save haproxytech/kubernetes-ingress:3.1.14 -o haproxy-ingress.tar
[ ] minikube image load haproxy-ingress.tar
[ ] minikube image ls | grep haproxy
[ ] helm install haproxy-ingress haproxytech/kubernetes-ingress \
 --namespace haproxy-controller \
 --create-namespace \
 --set controller.kind=Deployment \
 --set controller.ingressClass=haproxy \
 --set controller.service.type=NodePort \
 --set controller.image.repository=haproxytech/kubernetes-ingress \
 --set controller.image.tag=3.1.14
[ ] kubectl get pods -n haproxy-controller
[ ] kubectl patch deployment haproxy-ingress-kubernetes-ingress -n haproxy-controller -p='
{
  "spec": {
    "template": {
      "spec": {
        "terminationGracePeriodSeconds": 180,
        "containers": [
          {
            "name": "kubernetes-ingress-controller",
            "lifecycle": {
              "preStop": {
                "exec": {
                  "command": ["/bin/sh", "-c", "sleep 2 && kill -USR1 $(pidof haproxy) && sleep 30"]
                }
              }
            }
          }
        ]
      }
    }
  }
}'
[ ] k apply -f k8s/haproxy-ingress.yaml
[ ] echo "$(minikube ip) haproxy.local" | sudo tee -a /etc/hosts
[ ] curl "http://haproxy.local:$(kubectl get svc -n haproxy-controller haproxy-ingress-kubernetes-ingress -o jsonpath='{.spec.ports[0].nodePort}')"


# To deploy new changes after code updates:
[ ] docker build -t ws-app ./go
[ ] minikube image load ws-app
[ ] minikube image ls | grep ws
[ ] k apply -f k8s/deployment.yaml

# To restart HAProxy Ingress
[ ] kubectl rollout restart deployment haproxy-ingress-kubernetes-ingress -n haproxy-controller