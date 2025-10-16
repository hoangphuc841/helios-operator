# 📚 Helios Operator Documentation

Welcome to the comprehensive documentation for Helios Operator! This is your central hub for everything you need to know about deploying, using, and contributing to Helios.

## 🎯 **Getting Started**

New to Helios? Start here for a smooth onboarding experience:

### 👥 **For Application Developers**

If you want to deploy applications using Helios:

1. **[🚀 Getting Started Guide](user-guide/01-getting-started.md)** - Complete setup in 15 minutes
2. **[📋 HeliosApp Specification](user-guide/02-helios-app-spec.md)** - Understand the configuration options
3. **[🔧 Troubleshooting](user-guide/03-troubleshooting.md)** - Solve common issues

### 👨‍💻 **For Platform Engineers**

If you're managing or contributing to the Helios platform:

1. **[🏗️ Architecture Overview](developer-guide/01-architecture.md)** - System design and components
2. **[⚙️ Development Setup](developer-guide/02-development-setup.md)** - Local development environment
3. **[🧪 Testing Guide](developer-guide/03-testing.md)** - Comprehensive testing strategies
4. **[🛠️ Makefile Guide](developer-guide/04-makefile-guide.md)** - Optimized Makefile for IDP workflows

## 📖 **Documentation Structure**

### 👥 **User Guide** - For Application Developers

| Document                                                       | Description                                       | Time to Read |
| -------------------------------------------------------------- | ------------------------------------------------- | ------------ |
| [🚀 Getting Started](user-guide/01-getting-started.md)         | Complete setup and first application deployment   | 15 minutes   |
| [📋 HeliosApp Specification](user-guide/02-helios-app-spec.md) | Detailed explanation of all configuration options | 10 minutes   |
| [🔧 Troubleshooting](user-guide/03-troubleshooting.md)         | Common issues and their solutions                 | 5 minutes    |

### 👨‍💻 **Developer Guide** - For Platform Engineers

| Document                                                        | Description                              | Time to Read |
| --------------------------------------------------------------- | ---------------------------------------- | ------------ |
| [🏗️ Architecture Overview](developer-guide/01-architecture.md)  | System design, components, and data flow | 20 minutes   |
| [⚙️ Development Setup](developer-guide/02-development-setup.md) | Local development environment and tools  | 30 minutes   |
| [🧪 Testing Guide](developer-guide/03-testing.md)               | Unit, integration, and E2E testing       | 15 minutes   |
| [🛠️ Makefile Guide](developer-guide/04-makefile-guide.md)       | Build automation and development tasks   | 10 minutes   |

### 📖 **Reference** - Technical Details

| Document                                                        | Description                      | Use Case              |
| --------------------------------------------------------------- | -------------------------------- | --------------------- |
| [⚙️ Helm Chart Values](reference/helm-chart-values.md)          | Complete configuration reference | Production deployment |
| [📊 Prometheus Metrics](reference/prometheus-metrics.md)        | Available metrics and monitoring | Observability setup   |
| [🔒 Pod Security Standards](security/pod-security-standards.md) | Security implementation details  | Security compliance   |

## 🎯 **Quick Navigation by Use Case**

### 🚀 **I want to deploy my first application**

→ [Getting Started Guide](user-guide/01-getting-started.md)

### ⚙️ **I need to configure Helios for production**

→ [Helm Chart Values](reference/helm-chart-values.md)

### 🔧 **Something isn't working**

→ [Troubleshooting Guide](user-guide/03-troubleshooting.md)

### 🏗️ **I want to understand how it works**

→ [Architecture Overview](developer-guide/01-architecture.md)

### 👨‍💻 **I want to contribute to the project**

→ [Development Setup](developer-guide/02-development-setup.md)

### 📊 **I need to set up monitoring**

→ [Prometheus Metrics](reference/prometheus-metrics.md)

### 🔒 **I need to understand security requirements**

→ [Pod Security Standards](security/pod-security-standards.md)

## 🌟 **Key Concepts**

### **HeliosApp**

The core Custom Resource that defines your application. It specifies source code, build pipeline, and deployment configuration with comprehensive validation and status reporting.

### **Automatic Pipeline Generation**

Helios automatically creates Tekton Pipelines from templates, eliminating the need to manage Pipeline resources manually. The system now includes robust error handling and retry logic.

### **GitOps Workflow**

Pure GitOps approach where ArgoCD manages deployments based on Git repository changes with enhanced sync status monitoring.

### **Real-time Status Updates**

Kubernetes Watches provide instant feedback on build status, deployment health, and sync status with structured logging and comprehensive metrics.

### **Clean Architecture**

The project follows Go best practices with:

- **Modular Design**: Separated concerns with `internal/reconciler/`, `internal/resources/`, `internal/common/`
- **Comprehensive Testing**: Unit tests with >70% coverage and integration tests
- **Error Handling**: Robust error handling with proper logging and retry logic
- **Code Quality**: Linting compliance and clean code structure

## 🔗 **External Resources**

- **[GitHub Repository](https://github.com/hoangphuc841/helios-operator)** - Source code and issues
- **[Helm Chart](https://github.com/hoangphuc841/helios-operator/tree/main/helm/helios-operator)** - Production deployment
- **[Tekton Documentation](https://tekton.dev/docs/)** - Understanding Tekton Pipelines
- **[ArgoCD Documentation](https://argo-cd.readthedocs.io/)** - Understanding ArgoCD

## 📞 **Support**

- **Issues**: [GitHub Issues](https://github.com/hoangphuc841/helios-operator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hoangphuc841/helios-operator/discussions)
- **Contributing**: [Contributing Guide](../CONTRIBUTING.md)

---

**Need help?** Check our [Troubleshooting Guide](user-guide/03-troubleshooting.md) or open an [issue](https://github.com/hoangphuc841/helios-operator/issues) 🆘
