{{- define "neuralops.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "neuralops.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "neuralops.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "neuralops.labels" -}}
app.kubernetes.io/name: {{ include "neuralops.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: neuralops
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}

{{- define "neuralops.namespace" -}}
{{- default .Values.global.namespace .Release.Namespace -}}
{{- end -}}
