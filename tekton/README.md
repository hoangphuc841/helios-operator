Helios CI/CD Tekton Pipeline
This guide provides step-by-step instructions to set up and run a complete "From Code to Cluster" CI/CD pipeline using Tekton. The pipeline automates building a NodeJS application, pushing it as a Docker image to Docker Hub, and updating a GitOps repository with the new image tag.
Prerequisites
Before starting, ensure you have the following:

A running Kubernetes cluster (e.g., Minikube, Kind, or Docker Desktop).
kubectl command-line tool installed and configured to connect to your cluster.
A Docker Hub account with a personal access token.
A GitHub account with a personal access token (with repo scope).
Two separate GitHub repositories:
Application Source Repo: Contains the NodeJS application code.
GitOps Repo: Contains the Kubernetes deployment.yaml manifest.



Step 1: Install Tekton and tkn CLI
Set up Tekton Pipelines on your cluster and install the tkn command-line tool for easier pipeline management.
Install tkn CLI
# Download the latest release
curl -LO https://github.com/tektoncd/cli/releases/download/v0.38.0/tkn_0.38.0_Linux_x86_64.tar.gz
# Extract the archive
tar -xvzf tkn_0.38.0_Linux_x86_64.tar.gz
# Move the binary to your path
sudo mv tkn /usr/local/bin/

Install Tekton Pipelines
# Apply the latest Tekton pipeline components to your cluster
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

(Optional) Install Tekton Dashboard
The Tekton Dashboard provides a graphical interface to view your pipelines.
# Install the Tekton Dashboard
kubectl apply --filename https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml
# Access the dashboard at http://localhost:9097
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097

Wait for all pods in the tekton-pipelines namespace to be in the Running state before proceeding.
kubectl get pods --namespace tekton-pipelines --watch

Step 2: Create Kubernetes Secrets
The pipeline requires credentials to push to Docker Hub and your GitOps repository. Store these securely as Kubernetes Secrets. 
Docker Hub Secret (You can edit directly dockerhub secrets and skips the below command in manual-pipelinerun.yaml)
kubectl create secret docker-registry docker-credentials \
  --docker-username=<YOUR_DOCKER_HUB_USERNAME> \
  --docker-password=<YOUR_DOCKER_HUB_ACCESS_TOKEN>

GitHub Secret
kubectl create secret generic git-credentials \
  --from-literal=token=<YOUR_GITHUB_PERSONAL_ACCESS_TOKEN>

Step 3: Apply Tekton Resources
Apply the necessary Tekton Tasks, Pipelines, and ServiceAccount to your cluster. Ensure all required YAML files (service-account.yaml, task-git-clone.yaml, task-kaniko-build.yaml, task-git-update-manifest.yaml, and pipeline.yaml) are in your current directory.
# Apply the ServiceAccount that links to the secrets
kubectl apply -f service-account.yaml
# Apply the individual Task definitions
kubectl apply -f task-git-clone.yaml
kubectl apply -f task-kaniko-build.yaml
kubectl apply -f task-git-update-manifest.yaml
# Apply the main Pipeline that connects the tasks
kubectl apply -f pipeline.yaml

Step 4: Configure and Run the Pipeline
Configure a PipelineRun to trigger the pipeline with the correct parameters.
Edit manual-pipelinerun.yaml
Modify the manual-pipelinerun.yaml file to point to your specific repositories and credentials. This configuration uses the secure ServiceAccount method.
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: final-run-1
spec:
  pipelineRef:
    name: from-code-to-cluster
  # Use the secrets configured via ServiceAccount
  serviceAccountName: pipeline-sa
  params:
    # Source Code Repo
    - name: app-repo-url
      value: "https://github.com/<YOUR_GITHUB_USERNAME>/simple-node-app.git"
    - name: app-repo-revision
      value: "main"
    # Docker Hub Image Repo
    - name: image-repo
      value: "docker.io/<YOUR_DOCKER_HUB_USERNAME>/simple-node-app"
    # GitOps Repo
    - name: gitops-repo-url
      value: "https://github.com/<YOUR_GITHUB_USERNAME>/helios-gitops.git"
    - name: manifest-path-in-gitops-repo
      value: "simple-node-app/deployment.yaml"
  # Provide separate storage folders for pipeline tasks
  workspaces:
    - name: source-workspace
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    - name: gitops-workspace
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi


Note: The pipeline uses the docker-credentials secret linked to the pipeline-sa ServiceAccount for secure authentication, replacing any direct login methods used during debugging.

Trigger the Pipeline
# Apply the file to start the pipeline
kubectl apply -f manual-pipelinerun.yaml

Monitor the Run
# Watch the pipeline logs in real-time
tkn pipelinerun logs -f final-run-1

Expected Outcome
Upon successful pipeline completion:

A new Docker image will be pushed to your Docker Hub repository.
A new commit, authored by "Tekton Pipeline", will appear in your GitOps repository, updating the deployment.yaml with the new image tag.
