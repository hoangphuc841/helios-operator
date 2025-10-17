{{/*
Expand the name of the chart.
*/}}
{{- define "helios-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "helios-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "helios-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "helios-operator.labels" -}}
helm.sh/chart: {{ include "helios-operator.chart" . }}
{{ include "helios-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.globalLabels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "helios-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "helios-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "helios-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "helios-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "helios-operator.namespace" -}}
{{- .Values.namespace.name }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "helios-operator.annotations" -}}
{{- with .Values.globalAnnotations }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Image name
*/}}
{{- define "helios-operator.image" -}}
{{- $registry := .Values.image.registry | default "" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}
{{- end }}

{{/*
Environment variables
*/}}
{{- define "helios-operator.env" -}}
- name: OPERATOR_NAME
  value: "helios-operator"
- name: OPERATOR_NAMESPACE
  value: {{ include "helios-operator.namespace" . | quote }}
{{- if .Values.logging.level }}
- name: LOG_LEVEL
  value: {{ .Values.logging.level | quote }}
{{- end }}
{{- if .Values.logging.format }}
- name: LOG_FORMAT
  value: {{ .Values.logging.format | quote }}
{{- end }}
{{- if .Values.debug.enabled }}
- name: DEBUG
  value: "true"
{{- end }}
{{- with .Values.env }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Security context for containers
*/}}
{{- define "helios-operator.securityContext" -}}
{{- if .Values.securityContext }}
{{- toYaml .Values.securityContext }}
{{- else }}
allowPrivilegeEscalation: false
capabilities:
  drop:
  - ALL
readOnlyRootFilesystem: true
runAsNonRoot: true
runAsUser: 65532
{{- end }}
{{- end }}

{{/*
Pod security context
*/}}
{{- define "helios-operator.podSecurityContext" -}}
{{- if .Values.podSecurityContext }}
{{- toYaml .Values.podSecurityContext }}
{{- else }}
fsGroup: 65532
runAsNonRoot: true
runAsUser: 65532
seccompProfile:
  type: RuntimeDefault
{{- end }}
{{- end }}
