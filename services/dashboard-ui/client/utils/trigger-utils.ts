import type { TTrigger } from '@/types'

export const shellQuote = (value: string) =>
  `'${value.replaceAll("'", `'"'"'`)}'`

export const filterTriggers = (
  triggers: TTrigger[],
  {
    source,
    authType,
    envelope,
  }: { source?: string; authType?: string; envelope?: string }
): TTrigger[] =>
  triggers.filter(
    (trigger) =>
      (!source ||
        (source === 'GitHub' && trigger?.preset === 'github') ||
        (source === 'Slack' && trigger?.preset === 'slack-events') ||
        (source === 'Datadog' && trigger?.preset === 'datadog') ||
        (source === 'AWS' && trigger?.preset?.startsWith('aws-')) ||
        (source === 'GCP' && trigger?.preset === 'google-pubsub') ||
        (source === 'Azure' && trigger?.preset === 'azure-event-grid') ||
        (source === 'Custom' &&
          (!trigger?.preset || trigger.preset === 'custom'))) &&
      (!authType || trigger?.auth_type === authType) &&
      (!envelope || trigger?.envelope === envelope)
  )

export const parseEventSelectorPath = (
  path: string | undefined
): Array<string | number> | null => {
  if (!path?.startsWith('$')) return null
  const segments: Array<string | number> = []
  let index = 1

  while (index < path.length) {
    if (path[index] === '.') {
      index += 1
      const match = path.slice(index).match(/^[A-Za-z_][A-Za-z0-9_-]*/)
      if (!match) return null
      segments.push(match[0])
      index += match[0].length
      continue
    }

    if (path[index] !== '[') return null
    index += 1
    const quote = path[index]
    if (quote === "'" || quote === '"') {
      index += 1
      let name = ''
      let closed = false
      while (index < path.length) {
        const character = path[index]
        if (character === '\\') {
          index += 1
          if (index >= path.length) return null
          name += path[index]
          index += 1
        } else if (character === quote) {
          index += 1
          closed = true
          break
        } else {
          name += character
          index += 1
        }
      }
      if (!closed || !name || path[index] !== ']') return null
      segments.push(name)
      index += 1
      continue
    }

    const match = path.slice(index).match(/^\d+/)
    if (!match) return null
    index += match[0].length
    if (path[index] !== ']') return null
    segments.push(Number(match[0]))
    index += 1
  }

  return segments
}

export const setExampleValue = (
  payload: Record<string, unknown>,
  path: string | undefined,
  value: string
) => {
  const segments = parseEventSelectorPath(path)
  if (!segments?.length || typeof segments[0] === 'number') return

  let current: Record<string, unknown> | unknown[] = payload
  for (let index = 0; index < segments.length - 1; index += 1) {
    const segment = segments[index]
    const next = typeof segments[index + 1] === 'number' ? [] : {}
    current[segment as keyof typeof current] = next
    current = next
  }
  current[segments.at(-1)! as keyof typeof current] = value
}

export const utf8Base64 = (value: string) => {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary)
}

const examplePayload = (trigger: TTrigger, eventType: string) => {
  if (trigger?.preset === 'slack-events')
    return {
      type: 'event_callback',
      team_id: 'T0123456789',
      api_app_id: 'A0123456789',
      event_id: 'Ev0123456789',
      event_time: 1785254400,
      event: { type: eventType, text: 'Example event' },
    }
  if (trigger?.preset === 'azure-event-grid')
    return [
      {
        id: 'example-event-1',
        eventType,
        subject: 'nuon/example',
        eventTime: '2026-07-28T12:00:00Z',
        dataVersion: '1.0',
        data: { example: 'value' },
      },
    ]
  if (trigger?.envelope === 'cloudevents')
    return {
      specversion: '1.0',
      id: 'example-event-1',
      source: 'manual-example',
      type: eventType,
      datacontenttype: 'application/json',
      data: { example: 'value' },
    }

  const payload: Record<string, unknown> = { data: { example: 'value' } }
  setExampleValue(payload, trigger?.id_from?.payload, 'example-event-1')
  setExampleValue(payload, trigger?.type_from?.payload, eventType)

  if (trigger?.envelope === 'pubsub_push')
    return {
      message: {
        messageId: 'example-event-1',
        data: utf8Base64(JSON.stringify(payload)),
      },
    }
  return payload
}

const curlHeaders = (trigger: TTrigger, eventType: string) => {
  const headers = ['Content-Type: application/json']
  if (trigger?.envelope === 'none') {
    if (trigger?.id_from?.header)
      headers.push(`${trigger.id_from.header}: example-event-1`)
    if (trigger?.type_from?.header)
      headers.push(`${trigger.type_from.header}: ${eventType}`)
  }
  if (trigger?.auth_type === 'api_key')
    headers.push(
      `${trigger?.auth_config?.header || 'X-Nuon-API-Key'}: <API_KEY>`
    )
  if (trigger?.auth_type === 'bearer_jwt')
    headers.push('Authorization: Bearer <JWT>')
  return headers
}

const retriggerName = (name: string) => {
  const cleaned = `nuon-${name}`.replaceAll(/[^A-Za-z0-9._-]+/g, '-')
  return cleaned.slice(0, 64).replace(/-+$/, '') || 'nuon-trigger'
}

const azureRetriggerName = (name: string) => {
  const cleaned = `nuon-${name}`.replaceAll(/[^A-Za-z0-9-]+/g, '-')
  return cleaned.slice(0, 64).replace(/^-+|-+$/g, '') || 'nuon-trigger'
}

export type TTriggerSetupStep = { title: string; command: string }

export const buildTriggerSetupSteps = (
  trigger: TTrigger,
  { ingressUrl, secret }: { ingressUrl?: string; secret?: string }
): TTriggerSetupStep[] => {
  if (!ingressUrl) return []
  const url = shellQuote(ingressUrl)
  const preset = trigger?.preset

  if (preset === 'aws-eventbridge') {
    const authParameters = JSON.stringify({
      ApiKeyAuthParameters: {
        ApiKeyName: 'X-Nuon-API-Key',
        ApiKeyValue: secret || '<SECRET>',
      },
    })
    const retrigger = retriggerName(trigger?.name || '')
    return [
      {
        title: 'Create a connection that stores the API key',
        command: `connection_arn=$(aws events create-connection \\
  --name ${retrigger} \\
  --authorization-type API_KEY \\
  --auth-parameters ${shellQuote(authParameters)} \\
  --query ConnectionArn --output text)`,
      },
      {
        title: 'Create an API destination pointing at the ingress URL',
        command: `destination_arn=$(aws events create-api-destination \\
  --name ${retrigger} \\
  --connection-arn "$connection_arn" \\
  --invocation-endpoint ${url} \\
  --http-method POST \\
  --query ApiDestinationArn --output text)`,
      },
      {
        title: 'Create a rule matching the events to forward',
        command: `aws events put-rule \\
  --name ${retrigger} \\
  --event-pattern '{"source":["aws.s3"]}'`,
      },
      {
        title: 'Send matched events to the API destination',
        command: `aws events put-targets \\
  --rule ${retrigger} \\
  --targets "Id=nuon,Arn=$destination_arn,RoleArn=<EVENTBRIDGE_INVOKE_ROLE_ARN>"`,
      },
    ]
  }

  if (preset === 'aws-sns')
    return [
      {
        title: 'Subscribe the ingress URL to your SNS topic',
        command: `aws sns subscribe \\
  --topic-arn ${shellQuote(trigger?.auth_config?.topic_arn || '<TOPIC_ARN>')} \\
  --protocol https \\
  --notification-endpoint ${url}`,
      },
    ]

  if (preset === 'google-pubsub') {
    const identity = trigger?.auth_config?.expected_email
    const audience = trigger?.auth_config?.audience?.at(0)
    const lines = [
      `gcloud pubsub subscriptions create ${retriggerName(trigger?.name || '')} \\`,
      `  --topic '<TOPIC>' \\`,
      `  --push-endpoint ${url} \\`,
      `  --push-auth-service-account ${shellQuote(identity || '<SERVICE_ACCOUNT_EMAIL>')}`,
    ]
    if (audience) {
      lines[3] += ' \\'
      lines.push(`  --push-auth-token-audience ${shellQuote(audience)}`)
    }
    return [
      {
        title: 'Create a push subscription pointing at the ingress URL',
        command: lines.join('\n'),
      },
    ]
  }

  if (preset === 'azure-event-grid')
    return [
      {
        title: 'Create an Event Grid subscription pointing at the ingress URL',
        command: `az eventgrid event-subscription create \\
  --name ${azureRetriggerName(trigger?.name || '')} \\
  --source-resource-id '<SOURCE_RESOURCE_ID>' \\
  --endpoint ${url} \\
  --endpoint-type webhook \\
  --event-delivery-schema eventgridschema \\
  --max-events-per-batch 1 \\
  --delivery-attribute-mapping X-Nuon-API-Key static ${shellQuote(secret || '<SECRET>')} true`,
      },
    ]

  if (preset === 'slack-events')
    return [
      {
        title: 'Paste this Request URL into your Slack app',
        command: `# api.slack.com/apps → your app → Event Subscriptions → Request URL\n${ingressUrl}`,
      },
    ]

  if (preset === 'datadog') {
    const webhookName = retriggerName(trigger?.name || '')
    const payload = JSON.stringify({
      event_id: '$ID',
      event_type: '$EVENT_TYPE',
      alert_id: '$ALERT_ID',
      title: '$ALERT_TITLE',
      transition: '$ALERT_TRANSITION',
      scope: '$ALERT_SCOPE',
      message: '$EVENT_MSG',
      occurred_at: '$DATE_POSIX',
      link: '$LINK',
      tags: '$TAGS',
    })
    const request = JSON.stringify({
      name: webhookName,
      url: ingressUrl,
      encode_as: 'json',
      custom_headers: JSON.stringify({
        'X-Nuon-API-Key': secret || '<SECRET>',
      }),
      payload,
    })
    return [
      {
        title: 'Create the Datadog Webhooks integration',
        command: `curl --request POST \\
  --url "https://api.\${DD_SITE:-datadoghq.com}/api/v1/integration/webhooks/configuration/webhooks" \\
  --header "DD-API-KEY: $DD_API_KEY" \\
  --header "DD-APPLICATION-KEY: $DD_APP_KEY" \\
  --header 'Content-Type: application/json' \\
  --data ${shellQuote(request)}`,
      },
      {
        title: 'Notify the webhook from a monitor',
        command: `# Add this recipient to the monitor notification message:\n@webhook-${webhookName}`,
      },
    ]
  }

  if (preset === 'github')
    return [
      {
        title: 'Create the repository webhook',
        command: `gh api repos/<OWNER>/<REPO>/hooks \\
  --method POST \\
  -f name=web \\
  -f ${shellQuote(`config[url]=${ingressUrl}`)} \\
  -f 'config[content_type]=json' \\
  -f ${shellQuote(`config[secret]=${secret || '<SECRET>'}`)} \\
  -f 'events[]=push'`,
      },
    ]

  return []
}

export const buildTriggerCurl = (
  trigger: TTrigger,
  ingressUrl: string | undefined,
  eventType = 'example.event'
) => {
  if (!ingressUrl || trigger?.envelope === 'sns') return ''
  const body = JSON.stringify(examplePayload(trigger, eventType))
  const lines = ['curl --request POST', `  --url ${shellQuote(ingressUrl)}`]
  for (const header of curlHeaders(trigger, eventType))
    lines.push(`  --header ${shellQuote(header)}`)
  if (trigger?.auth_type === 'basic')
    lines.push(
      `  --user ${shellQuote(`${trigger?.auth_config?.username || 'nuon'}:<PASSWORD>`)}`
    )

  if (trigger?.auth_type === 'hmac') {
    const configuredAlgorithm = trigger?.auth_config?.algorithm
    const algorithm =
      configuredAlgorithm === 'sha256' || configuredAlgorithm === 'sha512'
        ? configuredAlgorithm
        : 'sha256'
    const encoding = trigger?.auth_config?.encoding || 'hex'
    const encode = encoding === 'base64' ? 'openssl base64 -A' : 'xxd -p -c 256'
    const prefix = trigger?.auth_config?.prefix || ''
    const header = trigger?.auth_config?.header || 'X-Nuon-Signature'
    const slack = trigger?.preset === 'slack-events'
    const requestHeaders = slack
      ? [
          `  --header 'X-Slack-Request-Timestamp: '"$timestamp"`,
          `  --header 'X-Slack-Signature: v0='"$signature"`,
        ]
      : [`  --header ${shellQuote(`${header}: ${prefix}`)}"$signature"`]
    const curl = [...lines, ...requestHeaders, '  --data "$body"'].join(' \\\n')
    const signatureInput = slack
      ? `'v0:%s:%s' "$timestamp" "$body"`
      : `'%s' "$body"`
    const slackPrelude = slack
      ? `TRIGGER_SECRET='<SLACK_SIGNING_SECRET>'\ntimestamp=$(date +%s)\n`
      : `TRIGGER_SECRET='<SECRET>'\n`
    return `body=${shellQuote(body)}\n${slackPrelude}signature=$(printf ${signatureInput} | openssl dgst -${algorithm} -hmac "$TRIGGER_SECRET" -binary | ${encode})\n${curl}`
  }

  lines.push(`  --data ${shellQuote(body)}`)
  return lines.join(' \\\n')
}
