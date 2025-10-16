# 🔧 Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using Helios Operator.

## 🔍 **Diagnostic Commands**

### Check Operator Status

```bash
# Verify operator is running
kubectl get pods -n system -l app.kubernetes.io/name=helios-operator

# Check operator logs
kubectl logs -n system deployment/helios-operator -f

# Verify CRD is installed
kubectl get crd heliosapps.platform.helios.io
```

### Check HeliosApp Status

```bash
# Get detailed status
kubectl describe heliosapp <app-name> -n <namespace>

# Check status conditions
kubectl get heliosapp <app-name> -n <namespace> -o yaml
```

## 🚨 **Common Issues & Solutions**

### 1. Operator Not Creating Resources

**Symptoms:**

- HeliosApp shows no status conditions
- No Tekton resources created
- No ArgoCD Application created

**Diagnosis:**

```bash
# Check operator logs for errors
kubectl logs -n system deployment/helios-operator | grep ERROR

# Verify RBAC permissions
kubectl describe clusterrole helios-operator-manager-role

# Check if CRD is installed
kubectl get crd heliosapps.platform.helios.io
```

**Solutions:**

```bash
# Reinstall CRDs
kubectl apply -f config/crd/bases/platform.helios.io_heliosapps.yaml

# Check RBAC
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml

# Restart operator
kubectl rollout restart deployment/helios-operator -n system
```

### 2. Tekton Pipeline Not Created

**Symptoms:**

- HeliosApp status shows "PipelineReady: False"
- No Pipeline resource found

**Diagnosis:**

```bash
# Check for Pipeline resources
kubectl get pipelines -n <namespace>

# Check operator logs for Pipeline creation errors
kubectl logs -n system deployment/helios-operator | grep -i pipeline

# Verify Tekton is installed
kubectl get pods -n tekton-pipelines
```

**Solutions:**

```bash
# Reinstall Tekton if missing
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

# Check RBAC for Tekton resources
kubectl get clusterrole helios-operator-manager-role -o yaml | grep -A 10 tekton

# Manually trigger reconciliation
kubectl annotate heliosapp <app-name> -n <namespace> trigger="$(date +%s)"
```

### 3. ArgoCD Application Not Created

**Symptoms:**

- No ArgoCD Application in `argocd` namespace
- HeliosApp status shows "ApplicationSynced: Unknown"

**Diagnosis:**

```bash
# Check ArgoCD Applications
kubectl get applications -n argocd

# Verify ArgoCD is running
kubectl get pods -n argocd

# Check operator logs for ArgoCD errors
kubectl logs -n system deployment/helios-operator | grep -i argocd
```

**Solutions:**

```bash
# Reinstall ArgoCD if missing
helm repo add argo https://argoproj.github.io/argo-helm
helm install argocd argo/argo-cd --namespace argocd

# Check RBAC for ArgoCD resources
kubectl get clusterrole helios-operator-manager-role -o yaml | grep -A 10 argoproj

# Verify GitOps repository access
kubectl describe application <app-name>-argocd -n argocd
```

### 4. Build Failures

**Symptoms:**

- PipelineRun fails
- Build logs show errors
- Image not pushed to registry

**Diagnosis:**

```bash
# Check PipelineRun status
kubectl get pipelinerun -n <namespace>

# View build logs
kubectl logs -n <namespace> -l tekton.dev/pipelineRun=<pipelinerun-name>

# Check ServiceAccount permissions
kubectl describe serviceaccount <service-account> -n <namespace>
```

**Solutions:**

```bash
# Verify container registry credentials
kubectl get secrets -n <namespace>

# Check PVC is available
kubectl get pvc <pvc-name> -n <namespace>

# Verify Git repository access
kubectl logs -n <namespace> -l tekton.dev/task=git-clone
```

### 5. ArgoCD Sync Issues

**Symptoms:**

- ArgoCD Application shows "OutOfSync"
- Deployment not updated
- Sync errors in ArgoCD UI

**Diagnosis:**

```bash
# Check ArgoCD Application status
kubectl describe application <app-name>-argocd -n argocd

# View ArgoCD logs
kubectl logs -n argocd deployment/argocd-application-controller

# Check GitOps repository
kubectl get application <app-name>-argocd -n argocd -o yaml
```

**Solutions:**

```bash
# Force sync in ArgoCD
kubectl patch application <app-name>-argocd -n argocd --type merge -p '{"operation":{"sync":{"syncStrategy":{"force":true}}}}'

# Check GitOps repository access
kubectl get secret argocd-repo-credentials -n argocd

# Verify manifests in GitOps repo
git clone <gitops-repo-url>
ls <gitops-path>
```

### 6. Webhook Not Triggering Builds

**Symptoms:**

- Git push doesn't trigger build
- EventListener not receiving events
- No PipelineRun created on push

**Diagnosis:**

```bash
# Check EventListener status
kubectl get eventlisteners -n <namespace>

# Check EventListener service
kubectl get svc -n <namespace> | grep eventlistener

# View EventListener logs
kubectl logs -n <namespace> -l app.kubernetes.io/name=eventlistener
```

**Solutions:**

```bash
# Verify webhook URL
kubectl get svc -n <namespace> | grep eventlistener

# Check webhook secret matches
kubectl get heliosapp <app-name> -n <namespace> -o jsonpath='{.spec.webhookSecret}'

# Test webhook manually
curl -X POST -H "X-GitHub-Event: push" -H "X-Hub-Signature-256: sha256=..." <webhook-url>
```

## 🔧 **Advanced Troubleshooting**

### Debug Mode

Enable debug logging in the operator:

```bash
# Patch operator deployment
kubectl patch deployment helios-operator -n system --type='merge' -p='{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'

# View debug logs
kubectl logs -n system deployment/helios-operator -f
```

### Resource Cleanup

Clean up stuck resources:

```bash
# Delete stuck PipelineRuns
kubectl delete pipelinerun --all -n <namespace>

# Reset HeliosApp status
kubectl patch heliosapp <app-name> -n <namespace> --type='merge' -p='{"status":{}}'

# Force delete resources
kubectl delete heliosapp <app-name> -n <namespace> --force --grace-period=0
```

### Network Issues

Check network connectivity:

```bash
# Test Git repository access
kubectl run test-git --rm -i --tty --image=alpine/git -- sh
git ls-remote https://github.com/your-org/your-repo.git

# Test container registry access
kubectl run test-registry --rm -i --tty --image=alpine -- sh
wget -O- https://docker.io/v2/

# Test ArgoCD connectivity
kubectl port-forward svc/argocd-server -n argocd 8080:443
curl -k https://localhost:8080/healthz
```

## 📊 **Monitoring & Metrics**

### Check Prometheus Metrics

```bash
# Port forward to metrics endpoint
kubectl port-forward -n system svc/helios-operator-metrics 8443:8443

# View metrics
curl -k https://localhost:8443/metrics
```

### Key Metrics to Monitor

- `heliosapp_reconciliation_total` - Total reconciliations
- `heliosapp_reconciliation_duration_seconds` - Reconciliation time
- `heliosapp_pipeline_status_total` - Pipeline success/failure rates
- `heliosapp_deployment_health_total` - Deployment health status

## 🆘 **Getting Help**

### Collect Debug Information

When reporting issues, collect this information:

```bash
# Create debug script
cat > debug-info.sh << 'EOF'
#!/bin/bash
echo "=== Helios Operator Debug Information ==="
echo "Date: $(date)"
echo ""

echo "=== Operator Status ==="
kubectl get pods -n system -l app.kubernetes.io/name=helios-operator
echo ""

echo "=== CRD Status ==="
kubectl get crd heliosapps.platform.helios.io
echo ""

echo "=== HeliosApp Resources ==="
kubectl get heliosapps -A
echo ""

echo "=== Tekton Resources ==="
kubectl get pipelines,eventlisteners,triggerbindings,triggertemplates -A
echo ""

echo "=== ArgoCD Applications ==="
kubectl get applications -n argocd
echo ""

echo "=== Operator Logs (last 50 lines) ==="
kubectl logs -n system deployment/helios-operator --tail=50
EOF

chmod +x debug-info.sh
./debug-info.sh > debug-info.txt
```

### Reporting Issues

When opening an issue on GitHub, include:

1. **Debug information** from the script above
2. **HeliosApp YAML** that's causing issues
3. **Expected vs actual behavior**
4. **Steps to reproduce**
5. **Environment details** (Kubernetes version, OS, etc.)

### Community Support

- **GitHub Issues**: [Report bugs and request features](https://github.com/hoangphuc841/helios-operator/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/hoangphuc841/helios-operator/discussions)
- **Documentation**: [Check our guides](../index.md)

## 🎯 **Prevention Tips**

### Best Practices

1. **Validate configurations** before applying
2. **Use descriptive names** for resources
3. **Monitor logs** regularly
4. **Keep dependencies updated**
5. **Test in staging** before production

### Regular Maintenance

```bash
# Check for outdated resources
kubectl get heliosapps -A --sort-by=.metadata.creationTimestamp

# Clean up old PipelineRuns
kubectl delete pipelinerun --field-selector=status.phase=Succeeded -A

# Verify all dependencies are healthy
kubectl get pods -n tekton-pipelines
kubectl get pods -n argocd
kubectl get pods -n system
```

---

**Still having issues?** Open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues) with your debug information 🆘
