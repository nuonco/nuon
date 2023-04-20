{{- define "common.classes.ingress" -}}

  {{- $fullName := include "common.fullname" . -}}
  {{- $values := .Values.ingress -}}

  {{- if hasKey . "ObjectValues" -}}
    {{- with .ObjectValues.ingress -}}
      {{- $values = . -}}
    {{- end -}}
  {{ end -}}

  {{- $primaryService := get .Values.service (include "common.service.primary" .) -}}

  {{- $defaultServiceName := $fullName -}}
  {{- if and (hasKey $primaryService "nameOverride") $primaryService.nameOverride -}}
    {{- $defaultServiceName = printf "%v-%v" $defaultServiceName $primaryService.nameOverride -}}
  {{- end -}}
  {{- $defaultServicePort := get $primaryService.ports (include "common.service.ports.primary" (dict "values" $primaryService)) -}}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "common.fullname" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
    {{- with $values.labels }}
       {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- with $values.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if $values.tls }}
  tls:
    {{- range $values.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ tpl . $ | quote }}
        {{- end }}
      {{- if .secretName }}
      secretName: {{ tpl .secretName $ | quote}}
      {{- end }}
    {{- end }}
  {{- end }}
  rules:
    {{- range $values.hosts }}
    - http:
        paths:
          {{- range .paths }}
          {{- $service := $defaultServiceName -}}
          {{- $port := $defaultServicePort.port -}}
          {{- if .service -}}
            {{- $service = default $service .service.name -}}
            {{- $port = default $port .service.port -}}
          {{- end }}
          - path: {{ tpl .path $ | quote }}
            pathType: {{ $.pathType | default "Prefix" }}
            backend:
              service:
                name: {{ $service }}
                port:
                  number: {{ $port }}
          {{- end }}
  {{- end }}
{{- end }}

