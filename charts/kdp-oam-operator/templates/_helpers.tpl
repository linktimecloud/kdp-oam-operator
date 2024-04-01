{{/*
Expand the name of the chart.
*/}}
{{- define "kdp-oam-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kdp-oam-operator.fullname" -}}
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
{{- define "kdp-oam-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kdp-oam-operator.labels" -}}
helm.sh/chart: {{ include "kdp-oam-operator.chart" . }}
{{ include "kdp-oam-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "kdp-oam-apiserver.labels" -}}
helm.sh/chart: {{ include "kdp-oam-operator.chart" . }}
{{ include "kdp-oam-apiserver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kdp-oam-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kdp-oam-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "kdp-oam-apiserver.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kdp-oam-operator.name" . }}-apiserver
app.kubernetes.io/instance: {{ .Release.Name }}-apiserver
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kdp-oam-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kdp-oam-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "kdp-oam-apiserver.serviceAccountName" -}}
{{- if .Values.apiserver.serviceAccount.create }}
{{- default (include "kdp-oam-operator.fullname" .) .Values.apiserver.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.apiserver.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the default kdp-oam-operator app version
*/}}
{{- define "kdp-oam-operator.defaultTag" -}}
  {{- default .Chart.AppVersion .Values.images.tag }}
{{- end -}}
