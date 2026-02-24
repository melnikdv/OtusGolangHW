{{/* Global validation of CHANGEME placeholders */}}
{{- define "chart.validateValues" -}}
  {{- $all := toYaml .Values -}}
  {{- if contains "CHANGEME" $all }}
    {{- fail "Validation failed: values.yaml contains placeholder strings with 'CHANGEME'. Please update all fields before deploying." }}
  {{- end }}
{{- end }}
