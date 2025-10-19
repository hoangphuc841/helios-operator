# Tekton Resources

This directory contains all Tekton Pipeline resources required by the Helios Operator.

## Files

- `pipeline.yaml` - Main CI/CD pipeline (from-code-to-cluster)
- `service-account.yaml` - ServiceAccount (pipeline-sa) for running pipelines
- `tekton-triggers-sa.yaml` - ServiceAccount (tekton-triggers-sa) for EventListener
- `workspace-pvc.yaml` - PersistentVolumeClaim (shared-workspace-pvc) for storage
- `task-*.yaml` - Individual Tekton Tasks

## Installation

```bash
kubectl apply -f tekton/
```
