# Pod Security Standards Implementation

This document describes the Pod Security Standards implementation in the Helios Operator project.

## Overview

The Helios Operator is configured to adhere to the **restricted** Pod Security Standards, which is the most secure level defined by Kubernetes. This ensures that all pods created by the operator meet the highest security requirements.

## Security Configuration

### Controller Pod Security

The controller manager pod is configured with the following security settings:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  fsGroup: 65532
  seccompProfile:
    type: RuntimeDefault

containers:
  - securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 65532
      capabilities:
        drop:
          - "ALL"
```

### Namespace Security Labels

The following namespaces are configured with restricted Pod Security Standards:

- `helios-operator-system`: Controller namespace
- `helios-apps`: Applications created by the operator
- `argocd`: ArgoCD namespace
- `tekton-pipelines`: Tekton Pipelines namespace
- `tekton-triggers`: Tekton Triggers namespace

Each namespace includes these labels:

```yaml
labels:
  pod-security.kubernetes.io/enforce: restricted
  pod-security.kubernetes.io/audit: restricted
  pod-security.kubernetes.io/warn: restricted
```

## Security Features

### 1. Non-Root Execution

- All containers run as non-root user (UID 65532)
- Prevents privilege escalation attacks

### 2. Read-Only Root Filesystem

- Container root filesystem is read-only
- Temporary directory mounted for writable files
- Prevents malicious file system modifications

### 3. Dropped Capabilities

- All Linux capabilities are dropped
- Prevents container from gaining additional privileges

### 4. Seccomp Profile

- Runtime default seccomp profile enabled
- Restricts system calls available to containers

### 5. No Privilege Escalation

- Privilege escalation is disabled
- Prevents containers from gaining root privileges

## Resource-Specific Security

### Tekton Pipelines

- Pipeline tasks run with restricted security contexts
- Service accounts with minimal required permissions
- No privileged containers or host access

### ArgoCD Applications

- Applications deployed with restricted security contexts
- Minimal RBAC permissions
- No host network or process access

## Security Best Practices

### 1. Regular Updates

- Keep Kubernetes cluster updated
- Update operator image regularly
- Monitor security advisories

### 2. Network Security

- Use NetworkPolicies to restrict traffic
- Enable TLS for all communications
- Use service mesh for advanced security

### 3. Image Security

- Scan container images for vulnerabilities
- Use minimal base images
- Sign images with trusted registries

### 4. Monitoring

- Monitor security events and violations
- Set up alerts for security policy violations
- Regular security audits

## Compliance

This implementation ensures compliance with:

- **Pod Security Standards**: Restricted level
- **CIS Kubernetes Benchmark**: Key security controls
- **NIST Cybersecurity Framework**: Protect and Detect functions
- **SOC 2**: Security and availability principles

## Troubleshooting

### Common Issues

1. **Pod Security Violations**

   - Check namespace labels
   - Verify security context configuration
   - Review container requirements

2. **Permission Denied Errors**

   - Ensure service accounts have required permissions
   - Check RBAC configuration
   - Verify file system permissions

3. **Container Startup Failures**
   - Check if application requires privileged access
   - Review security context requirements
   - Consider security policy exceptions if necessary

### Debugging Commands

```bash
# Check namespace security labels
kubectl get ns --show-labels

# Check pod security context
kubectl describe pod <pod-name> -n <namespace>

# Check security policy violations
kubectl get events --field-selector reason=FailedCreate

# Verify RBAC permissions
kubectl auth can-i <verb> <resource> --as=system:serviceaccount:<namespace>:<serviceaccount>
```

## References

- [Kubernetes Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Pod Security Policies](https://kubernetes.io/docs/concepts/security/pod-security-policy/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
