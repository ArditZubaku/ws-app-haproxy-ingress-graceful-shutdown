package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	println("Cleaning up existing resources...")
	execCmdIgnoreError("minikube delete --all --purge")
	execCmdIgnoreError("docker system prune -af")

	println("Checking prerequisites...")
	println("Docker version:", execCmdGetOutput("docker --version"))
	println("Minikube version:", execCmdGetOutput("minikube version --short"))

	println("Starting minikube...")
	execCmd("minikube start")

	println("Checking minikube status...")
	execCmd("minikube status")

	execCmd("docker build -t ws-app ./go")
	execCmd("minikube image load ws-app")
	execCmd("minikube image load ws-app")
	println("Image found in minikube: \n", execCmdGetOutput("minikube image ls | grep ws"))
	execCmd("kubectl apply -f k8s/deployment.yaml")
	execCmd("kubectl apply -f k8s/service.yaml")
	execCmd("helm repo add haproxytech https://haproxytech.github.io/helm-charts")
	execCmd("helm repo update")
	execCmd("docker pull haproxytech/kubernetes-ingress:3.1.14")
	execCmd("docker save haproxytech/kubernetes-ingress:3.1.14 -o haproxy-ingress.tar")
	execCmd("minikube image load haproxy-ingress.tar")
	println("Image found in minikube: \n", execCmdGetOutput("minikube image ls | grep haproxy"))
	execCmd("helm install haproxy-ingress haproxytech/kubernetes-ingress --namespace haproxy-controller --create-namespace --set controller.kind=Deployment --set controller.ingressClass=haproxy --set controller.service.type=NodePort --set controller.image.repository=haproxytech/kubernetes-ingress --set controller.image.tag=3.1.14")
	println("HAProxy-Controller pods: \n", execCmdGetOutput("kubectl get pods -n haproxy-controller"))
	execCmd(`kubectl patch deployment haproxy-ingress-kubernetes-ingress -n haproxy-controller -p='
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
}'`)
	execCmd("kubectl apply -f k8s/haproxy-ingress.yaml")

	// Configure /etc/hosts
	execCmd(`echo "$(minikube ip) haproxy.local" | sudo tee -a /etc/hosts`)

	// Debug: Check the service and get the port
	println("Debugging HAProxy service...")
	println("Services in haproxy-controller namespace:")
	println(execCmdGetOutput("kubectl get svc -n haproxy-controller"))

	nodePort := strings.TrimSpace(execCmdGetOutput(`kubectl get svc -n haproxy-controller haproxy-ingress-kubernetes-ingress -o jsonpath='{.spec.ports[0].nodePort}'`))
	println("HAProxy NodePort:", nodePort)

	// Get the ingress host programmatically
	ingressHost := strings.TrimSpace(execCmdGetOutput("kubectl get ingress haproxy-ingress -o jsonpath='{.spec.rules[0].host}'"))
	println("Ingress Host:", ingressHost)

	minikubeIP := strings.TrimSpace(execCmdGetOutput("minikube ip"))
	println("Minikube IP:", minikubeIP)

	// Wait a bit for services to be ready
	println("Waiting for HAProxy to be ready...")
	execCmd("kubectl wait --for=condition=available --timeout=60s deployment/haproxy-ingress-kubernetes-ingress -n haproxy-controller")

	// Try the curl using the dynamic ingress host
	curlURL := fmt.Sprintf("http://%s:%s", ingressHost, nodePort)
	println("Trying to curl:", curlURL)
	println("CURL RESULT:")
	println(execCmdGetOutput(fmt.Sprintf("curl -v %s", curlURL)))
}

func execCmd(cmd string) {
	fmt.Printf("Executing: %s\n", cmd)

	// Use shell for complex commands with pipes, quotes, etc.
	var execCmd *exec.Cmd
	if strings.Contains(cmd, "|") || strings.Contains(cmd, "$(") || strings.Contains(cmd, "'") {
		execCmd = exec.Command("bash", "-c", cmd)
	} else {
		parts := strings.Fields(cmd)
		execCmd = exec.Command(parts[0], parts[1:]...)
	}

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("ERROR executing command: %s\n", cmd)
		fmt.Printf("Error details: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("Successfully executed: %s\n", cmd)
}

func execCmdIgnoreError(cmd string) {
	fmt.Printf("Executing (ignore errors): %s\n", cmd)

	var execCmd *exec.Cmd
	if strings.Contains(cmd, "|") || strings.Contains(cmd, "$(") || strings.Contains(cmd, "'") {
		execCmd = exec.Command("bash", "-c", cmd)
	} else {
		parts := strings.Fields(cmd)
		execCmd = exec.Command(parts[0], parts[1:]...)
	}

	err := execCmd.Run()
	if err != nil {
		fmt.Printf("Command failed (ignored): %s - Error: %s\n", cmd, err.Error())
	} else {
		fmt.Printf("Successfully executed: %s\n", cmd)
	}
}

func execCmdGetOutput(cmd string) string {
	fmt.Printf("Executing (get output): %s\n", cmd)

	var execCmd *exec.Cmd
	if strings.Contains(cmd, "|") || strings.Contains(cmd, "$(") || strings.Contains(cmd, "'") {
		execCmd = exec.Command("bash", "-c", cmd)
	} else {
		parts := strings.Fields(cmd)
		execCmd = exec.Command(parts[0], parts[1:]...)
	}

	out, err := execCmd.Output()
	if err != nil {
		fmt.Printf("ERROR executing command: %s\n", cmd)
		fmt.Printf("Error details: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("Successfully executed: %s\n", cmd)
	return string(out)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
