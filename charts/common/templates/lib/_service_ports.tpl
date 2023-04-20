{{/*
Return the primary port for a given Service object.
*/}}
{{- define "common.service.ports.primary" -}}
  {{- $enabledPorts := dict -}}
  {{- range $name, $port := .values.ports -}}
    {{- if $port.enabled -}}
      {{- $_ := set $enabledPorts $name . -}}
    {{- end -}}
  {{- end -}}

  {{- if eq 0 (len $enabledPorts) }}
    {{- fail (printf "No ports are enabled for service \"%s\"!" .serviceName) }}
  {{- end }}

  {{- $result := "" -}}
  {{- range $name, $port := $enabledPorts -}}
    {{- if and (hasKey $port "primary") $port.primary -}}
      {{- $result = $name -}}
    {{- end -}}
  {{- end -}}

  {{- if not $result -}}
    {{- $result = keys $enabledPorts | first -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}



{{/*
Return the primary service object
*/}}
{{- define "common.service.primary" -}}
  {{- $enabledServices := dict -}}
  {{- range $name, $service := .Values.service -}}
    {{- if $service.enabled -}}
      {{- $_ := set $enabledServices $name . -}}
    {{- end -}}
  {{- end -}}

  {{- $result := "" -}}
  {{- range $name, $service := $enabledServices -}}
    {{- if and (hasKey $service "primary") $service.primary -}}
      {{- $result = $name -}}
    {{- end -}}
  {{- end -}}

  {{- if not $result -}}
    {{- $result = keys $enabledServices | first -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}
