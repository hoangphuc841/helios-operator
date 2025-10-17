# Basic Web Application Example

This example demonstrates how to deploy a simple web application using Helios Operator with minimal configuration.

## Overview

This example creates a basic web application that:

- Builds from a Git repository
- Deploys to Kubernetes via ArgoCD
- Uses default resource settings
- Includes basic health checks

## Prerequisites

- Helios Operator installed
- ArgoCD installed and running
- Tekton Pipelines and Triggers installed
- Access to a container registry

## Application Structure

```
my-web-app/
├── Dockerfile
├── package.json
├── src/
│   └── index.js
└── README.md
```

### Dockerfile

```dockerfile
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY src/ ./src/

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# Start application
CMD ["node", "src/index.js"]
```

### package.json

```json
{
  "name": "my-web-app",
  "version": "1.0.0",
  "description": "A simple web application",
  "main": "src/index.js",
  "scripts": {
    "start": "node src/index.js",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
```

### src/index.js

```javascript
const express = require("express");
const app = express();
const port = process.env.PORT || 3000;

// Health check endpoint
app.get("/health", (req, res) => {
  res
    .status(200)
    .json({ status: "healthy", timestamp: new Date().toISOString() });
});

// Main endpoint
app.get("/", (req, res) => {
  res.json({
    message: "Hello from my-web-app!",
    version: "1.0.0",
    environment: process.env.NODE_ENV || "development",
  });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
```

## HeliosApp Configuration

Create a file named `my-web-app.yaml`:

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-web-app
  namespace: default
  labels:
    app.kubernetes.io/name: my-web-app
    app.kubernetes.io/version: "1.0.0"
spec:
  # Source repository
  gitRepo: "https://github.com/your-username/my-web-app"
  gitBranch: "main"

  # Container image
  imageRepo: "your-registry.com/my-web-app"
  imageTag: "latest"

  # Application configuration
  port: 3000
  serviceAccount: "default"

  # Webhook configuration
  webhookSecret: "my-web-app-webhook-secret"

  # GitOps configuration
  gitopsRepo: "https://github.com/your-username/my-web-app-manifests"
  gitopsPath: "apps/my-web-app"
  gitopsBranch: "main"

  # Health check
  healthCheck:
    path: "/health"
    port: 3000
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
```

## GitOps Manifests

Create the following files in your GitOps repository at `apps/my-web-app/`:

### kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml

commonLabels:
  app.kubernetes.io/name: my-web-app
  app.kubernetes.io/version: "1.0.0"
  app.kubernetes.io/managed-by: helios-operator
```

### deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-web-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: my-web-app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: my-web-app
    spec:
      containers:
        - name: my-web-app
          image: your-registry.com/my-web-app:latest
          ports:
            - containerPort: 3000
              name: http
          env:
            - name: NODE_ENV
              value: "production"
            - name: PORT
              value: "3000"
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
```

### service.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-web-app
spec:
  selector:
    app.kubernetes.io/name: my-web-app
  ports:
    - name: http
      port: 80
      targetPort: 3000
      protocol: TCP
  type: ClusterIP
```

### ingress.yaml

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-web-app
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: my-web-app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-web-app
                port:
                  number: 80
```

## Deployment Steps

### 1. Create the HeliosApp

```bash
kubectl apply -f my-web-app.yaml
```

### 2. Monitor the Status

```bash
# Watch the HeliosApp status
kubectl get heliosapps -w

# Check detailed status
kubectl describe heliosapp my-web-app

# View events
kubectl get events --field-selector involvedObject.name=my-web-app
```

### 3. Verify ArgoCD Application

```bash
# Check if ArgoCD Application was created
kubectl get applications -n argocd

# Check ArgoCD Application status
kubectl describe application my-web-app -n argocd
```

### 4. Verify Tekton Resources

```bash
# Check EventListener
kubectl get eventlisteners

# Check TriggerBinding
kubectl get triggerbindings

# Check TriggerTemplate
kubectl get triggertemplates
```

### 5. Test the Application

```bash
# Port forward to test locally
kubectl port-forward svc/my-web-app 8080:80

# Test the application
curl http://localhost:8080/

# Test health endpoint
curl http://localhost:8080/health
```

## Expected Results

After successful deployment, you should see:

1. **HeliosApp Status**: `Ready` with all conditions `True`
2. **ArgoCD Application**: `Synced` and `Healthy`
3. **Deployment**: Running with 1 replica
4. **Service**: Available and accessible
5. **Ingress**: Configured (if ingress controller is available)

## Troubleshooting

### Common Issues

#### 1. Build Pipeline Fails

**Symptoms**: PipelineRun fails with build errors

**Solutions**:

- Check container registry credentials
- Verify Dockerfile syntax
- Ensure source repository is accessible

```bash
# Check PipelineRun logs
kubectl logs -l tekton.dev/pipelineRun=<pipeline-run-name>
```

#### 2. ArgoCD Application Not Syncing

**Symptoms**: ArgoCD Application shows `OutOfSync` status

**Solutions**:

- Verify GitOps repository access
- Check manifest syntax
- Ensure ArgoCD has proper permissions

```bash
# Check ArgoCD Application logs
kubectl logs -n argocd deployment/argocd-application-controller
```

#### 3. Application Not Accessible

**Symptoms**: Service is running but not accessible

**Solutions**:

- Check ingress controller configuration
- Verify service selector matches pod labels
- Check network policies

```bash
# Check service endpoints
kubectl get endpoints my-web-app

# Check pod labels
kubectl get pods --show-labels
```

## Next Steps

Once your basic web application is running:

1. **Add Monitoring**: Set up Prometheus metrics and Grafana dashboards
2. **Implement CI/CD**: Configure automated testing and deployment
3. **Add Security**: Implement RBAC, network policies, and secret management
4. **Scale**: Configure horizontal pod autoscaling
5. **Multi-Environment**: Deploy to staging and production environments

## Related Examples

- [Microservice with Database](../02-microservice-with-database.md)
- [Multi-Environment Deployment](../03-multi-environment.md)
- [Production-Ready Application](../04-production-ready.md)
