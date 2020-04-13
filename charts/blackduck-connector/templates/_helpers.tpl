{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ops.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ops.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ops.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ops.labels" -}}
helm.sh/chart: {{ include "ops.chart" . }}
{{ include "ops.selectorLabels" . }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "ops.selectorLabels" -}}
app: opssight
name: {{ .Release.Name }}
version: {{ .Values.imageTag }}
{{- end -}}

{{/*
Common labels without version
*/}}
{{- define "ops.labelsWithoutVersion" -}}
helm.sh/chart: {{ include "ops.chart" . }}
{{ include "ops.selectorLabelsWithoutVersion" . }}
{{- end -}}

{{/*
Selector labels without version
*/}}
{{- define "ops.selectorLabelsWithoutVersion" -}}
app: opssight
name: {{ .Release.Name }}
{{- end -}}

{{/*
Image pull secrets to pull the image
*/}}
{{- define "ops.imagePullSecrets" }}
{{- with .Values.imagePullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{/*
Add secured registries
*/}}
{{- define "ops.securedRegistries" -}}
{ {{ range $index, $element := .Values.securedRegistries }}
  {{- if $index -}},{{end}}
  {{ .url | quote -}}:
    {
        "url":{{ .url | quote }},
        "user":{{ .user | quote }},
        "password":{{ .password | quote }},
        "token":{{ .token | quote }}
    }
    {{- end }}
}
{{- end -}}

{{/*
Add external Black Duck
*/}}
{{- define "ops.externalBlackDuck" -}}
{ {{ range $index, $element := .Values.externalBlackDuck }}
  {{- if $index -}},{{end}}
  {{ .domain | quote -}}:
    {
        "scheme":{{ .scheme | quote }},
        "domain":{{ .domain | quote }},
        "port":{{ .port }},
        "user":{{ .user | quote }},
        "password":{{ .password | quote }},
        "concurrentScanLimit":{{ .concurrentScanLimit }}
    }
    {{- end }}
}
{{- end -}}
