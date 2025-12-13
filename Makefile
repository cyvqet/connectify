# ===== Colors =====
GREEN  := \033[32m
YELLOW := \033[33m
BLUE   := \033[34m
RED    := \033[31m
CYAN   := \033[36m
BOLD   := \033[1m
RESET  := \033[0m

# ===== Log helpers =====
INFO    = echo "$(BLUE)[INFO]$(RESET)"
SUCCESS = echo "$(GREEN)[SUCCESS]$(RESET)"
WARN    = echo "$(YELLOW)[WARN]$(RESET)"
ERROR   = echo "$(RED)[ERROR]$(RESET)"
STEP    = echo "$(CYAN)==>$(RESET)"

# Define phony targets (not real files, always execute commands)
.PHONY: docker docker-clean docker-deploy mysql-deploy mysql-clean redis-deploy redis-clean ingress-deploy ingress-clean redeploy all

# Docker image build - compile Go program and build Docker image
docker:
	@$(STEP) "$(BOLD)Step 1$(RESET): Removing old connectify binary..."
	@rm -f connectify || true

	@$(INFO) "Tidying Go module dependencies..."
	@go mod tidy

	@$(STEP) "$(BOLD)Step 2$(RESET): Cross-compiling Go program (Linux ARM)..."
	@GOOS=linux GOARCH=arm go build -tags=k8s -o connectify .
	# GOOS=linux: target operating system
	# GOARCH=arm: target CPU architecture
	# -tags=k8s: enable k8s build tag (conditional compilation)
	# -o connectify: output binary name

	@$(INFO) "Removing old Docker image (if exists)..."
	@docker rmi -f cyvqet/connectify:v1.0 2>/dev/null || true

	@$(STEP) "$(BOLD)Step 3$(RESET): Building new Docker image..."
	@docker build -t cyvqet/connectify:v1.0 .
	# -t: tag the image (username/image:version)
	# . : use current directory as build context (must contain Dockerfile)

	@$(INFO) "Cleaning dangling Docker images..."
	@docker image prune -f

	@$(INFO) "Cleaning build artifacts..."
	@rm -f connectify

	@$(SUCCESS) "Docker image build completed"

# Clean Kubernetes resources - remove deployed services and deployments
docker-clean:
	@$(INFO) "Cleaning Kubernetes resources..."
	@kubectl delete service connectify-record 2>/dev/null || true
	@kubectl delete deployment connectify-record-service 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "Kubernetes resources cleanup completed"

# Deploy application to Kubernetes
docker-deploy:
	@$(INFO) "Deploying application to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-service.yaml

	@$(INFO) "Waiting for Pods to start..."
	@sleep 5
	@kubectl get pods -l app=connectify-record

	@$(SUCCESS) "Application deployed successfully"

# Deploy MySQL to Kubernetes - create PV, PVC, Deployment, and Service
mysql-deploy:
	@$(INFO) "Deploying MySQL to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-mysql-pv.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-pvc.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-mysql-service.yaml

	@$(INFO) "Waiting for MySQL Pod to start..."
	@sleep 10
	@kubectl get pods -l app=connectify-record-mysql > /dev/null || true
	@kubectl get pv,pvc > /dev/null || true

	@$(INFO) "Creating database: connectify..."
	@kubectl exec -it $$(kubectl get pods -l app=connectify-record-mysql -o jsonpath='{.items[0].metadata.name}') -- \
		mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS connectify;"

	@$(SUCCESS) "MySQL deployed and database ready"

# Clean MySQL Kubernetes resources
mysql-clean:
	@$(INFO) "Cleaning MySQL Kubernetes resources..."
	@kubectl delete service connectify-record-mysql 2>/dev/null || true
	@kubectl delete deployment connectify-record-mysql 2>/dev/null || true
	@kubectl delete pvc connectify-mysql-pvc 2>/dev/null || true
	@kubectl delete pv connectify-mysql-pv 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "MySQL cleanup completed"

# Deploy Redis to Kubernetes
redis-deploy:
	@$(INFO) "Deploying Redis to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-redis-deployment.yaml
	@kubectl apply -f deploy/k8s/connectify-redis-service.yaml

	@$(INFO) "Waiting for Redis Pod to start..."
	@sleep 5
	@kubectl get pods -l app=connectify-record-redis

	@$(SUCCESS) "Redis deployed successfully"

# Clean Redis Kubernetes resources
redis-clean:
	@$(INFO) "Cleaning Redis Kubernetes resources..."
	@kubectl delete service connectify-record-redis 2>/dev/null || true
	@kubectl delete deployment connectify-record-redis 2>/dev/null || true
	@$(INFO) "Waiting for cleanup to complete..."
	@sleep 2
	@$(SUCCESS) "Redis cleanup completed"

# Deploy Ingress to Kubernetes
ingress-deploy:
	@$(INFO) "Deploying Ingress to Kubernetes..."
	@kubectl apply -f deploy/k8s/connectify-ingress.yaml
	@$(SUCCESS) "Ingress deployment completed"
	@$(INFO) "Access URL: http://localhost/"

# Clean Ingress resources
ingress-clean:
	@$(INFO) "Cleaning Ingress resources..."
	@kubectl delete ingress connectify-record-ingress 2>/dev/null || true
	@$(SUCCESS) "Ingress cleanup completed"

# Rebuild image and perform rolling update in Kubernetes
redeploy:
	@$(INFO) "Building new Docker image..."
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o connectify .
	@docker build -t cyvqet/connectify:v1.0 .

	@$(WARN) "Pushing image is commented out (enable if needed)"
	# @docker push cyvqet/connectify:v1.0

	@$(INFO) "Restarting Deployment (rolling update)..."
	@kubectl rollout restart deployment connectify-record-service

	@$(INFO) "Cleaning build artifacts..."
	@rm -f connectify

	@$(SUCCESS) "Rolling update completed successfully"

# One-click build, clean, and deploy - full CI/CD workflow
all: docker docker-clean mysql-deploy redis-deploy docker-deploy ingress-deploy
	@echo ""
	@echo "$(GREEN)==========================================$(RESET)"
	@echo "$(GREEN)  Deployment completed successfully!$(RESET)"
	@echo "$(GREEN)==========================================$(RESET)"
	@echo ""
	@$(INFO) "Access URL: http://localhost/test"
	@echo ""
	@$(INFO) "Common commands:"
	@echo "  kubectl get pods"
	@echo "  kubectl logs -f deployment/connectify-record-service"
	@echo "  kubectl logs -f -l app=connectify-record --all-containers"
	@echo "  kubectl rollout restart deployment/connectify-record-service"
	@echo "  kubectl exec -it \$$(kubectl get pods -l app=connectify-record -o jsonpath='{.items[0].metadata.name}') -- sh"
	@echo "  make docker-clean mysql-clean redis-clean ingress-clean"
	@echo ""
