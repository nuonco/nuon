import { describe, expect, test } from 'bun:test'
import type { TTrigger } from '@/types'
import {
  buildTriggerCurl,
  buildTriggerSetupSteps,
  filterTriggers,
  parseEventSelectorPath,
  setExampleValue,
  shellQuote,
  utf8Base64,
} from './trigger-utils'

describe('trigger-utils', () => {
  test('quotes malicious shell values without an unquoted semicolon', () => {
    const quoted = shellQuote("x'; touch /tmp/pwn; #")
    expect(quoted).toBe(`'x'"'"'; touch /tmp/pwn; #'`)
    expect(quoted).toEndWith(`'; touch /tmp/pwn; #'`)
  })

  test('quotes single quotes in event types and headers', () => {
    const curl = buildTriggerCurl(
      {
        envelope: 'none',
        type_from: { header: "X-Event'Type" },
        auth_type: 'api_key',
        auth_config: { header: "X-Key'Name" },
      },
      'https://example.com/ingress',
      "release'created"
    )
    expect(curl).toContain(`'X-Event'"'"'Type: release'"'"'created'`)
    expect(curl).toContain(`'X-Key'"'"'Name: <API_KEY>'`)
  })

  test('parses quoted names as one key', () => {
    expect(parseEventSelectorPath("$['build.version']")).toEqual([
      'build.version',
    ])
  })

  test('creates arrays for index selectors', () => {
    const payload: Record<string, unknown> = {}
    setExampleValue(payload, '$.a.b[0]', 'value')
    expect(payload).toEqual({ a: { b: ['value'] } })
  })

  test.each(['$..x', '$.a[-1]', '$[*]', "$['a','b']", 'garbage'])(
    'rejects invalid selector %s and skips the value',
    (path) => {
      const payload = { existing: true }
      expect(parseEventSelectorPath(path)).toBeNull()
      setExampleValue(payload, path, 'value')
      expect(payload).toEqual({ existing: true })
    }
  )

  test('falls back to sha256 for unsupported HMAC algorithms', () => {
    const trigger: TTrigger = {
      envelope: 'none',
      auth_type: 'hmac',
      auth_config: { algorithm: 'sha256; touch /tmp/pwn' },
    }
    const curl = buildTriggerCurl(trigger, 'https://example.com/ingress')
    expect(curl).toContain("TRIGGER_SECRET='<SECRET>'")
    expect(curl).toContain('openssl dgst -sha256 ')
    expect(curl).not.toContain('touch /tmp/pwn')
  })

  test('uses the CloudEvents source field in example requests', () => {
    const curl = buildTriggerCurl(
      { envelope: 'cloudevents', auth_type: 'none' },
      'https://example.com/ingress'
    )
    expect(curl).toContain('"source":"manual-example"')
    expect(curl).not.toContain('"trigger":"manual-example"')
  })

  test('encodes UTF-8 content to base64', () => {
    const value = JSON.stringify({ type: 'événement ✓' })
    const bytes = Uint8Array.from(atob(utf8Base64(value)), (character) =>
      character.charCodeAt(0)
    )
    expect(new TextDecoder().decode(bytes)).toBe(value)
  })

  describe('filterTriggers', () => {
    const triggers: TTrigger[] = [
      { id: 'evs-1', preset: 'github', auth_type: 'hmac', envelope: 'none' },
      {
        id: 'evs-2',
        preset: 'aws-eventbridge',
        auth_type: 'api_key',
        envelope: 'cloudevents',
      },
      {
        id: 'evs-3',
        preset: 'google-pubsub',
        auth_type: 'hmac',
        envelope: 'cloudevents',
      },
      { id: 'evs-4' },
      { id: 'evs-5', preset: 'aws-sns' },
      { id: 'evs-6', preset: 'azure-event-grid' },
      { id: 'evs-7', preset: 'slack-events' },
      { id: 'evs-8', preset: 'datadog' },
    ]

    test('returns all triggers when no filters are set', () => {
      expect(filterTriggers(triggers, {})).toEqual(triggers)
    })

    test('filters by auth type', () => {
      expect(
        filterTriggers(triggers, { authType: 'hmac' }).map((s) => s.id)
      ).toEqual(['evs-1', 'evs-3'])
    })

    test.each([
      ['GitHub', ['evs-1']],
      ['AWS', ['evs-2', 'evs-5']],
      ['GCP', ['evs-3']],
      ['Azure', ['evs-6']],
      ['Slack', ['evs-7']],
      ['Datadog', ['evs-8']],
      ['Custom', ['evs-4']],
    ])('filters by %s source', (source, expected) => {
      expect(filterTriggers(triggers, { source }).map((s) => s.id)).toEqual(
        expected
      )
    })

    test('filters by envelope', () => {
      expect(
        filterTriggers(triggers, { envelope: 'cloudevents' }).map((s) => s.id)
      ).toEqual(['evs-2', 'evs-3'])
    })

    test('combines auth type and envelope filters', () => {
      expect(
        filterTriggers(triggers, {
          source: 'GCP',
          authType: 'hmac',
          envelope: 'cloudevents',
        }).map((s) => s.id)
      ).toEqual(['evs-3'])
    })

    test('excludes triggers missing the filtered field', () => {
      expect(filterTriggers(triggers, { authType: 'none' })).toEqual([])
    })
  })

  describe('buildTriggerSetupSteps', () => {
    test('returns no steps without an ingress URL', () => {
      expect(
        buildTriggerSetupSteps({ name: 'x', preset: 'github' }, {})
      ).toEqual([])
    })

    test('returns no steps for triggers without a recognized preset', () => {
      expect(
        buildTriggerSetupSteps(
          { name: 'x', preset: 'custom' },
          { ingressUrl: 'https://example.com/ingress' }
        )
      ).toEqual([])
      expect(
        buildTriggerSetupSteps(
          { name: 'x' },
          { ingressUrl: 'https://example.com/ingress' }
        )
      ).toEqual([])
    })

    test('splits EventBridge setup into four titled steps', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'my trigger', preset: 'aws-eventbridge' },
        { ingressUrl: 'https://example.com/ingress', secret: "s3cr'et" }
      )
      expect(steps.map((step) => step.title)).toEqual([
        'Create a connection that stores the API key',
        'Create an API destination pointing at the ingress URL',
        'Create a rule matching the events to forward',
        'Send matched events to the API destination',
      ])
      expect(steps[0]?.command).toContain(
        shellQuote(
          '{"ApiKeyAuthParameters":{"ApiKeyName":"X-Nuon-API-Key","ApiKeyValue":"s3cr\'et"}}'
        )
      )
      expect(steps[0]?.command).toContain('--name nuon-my-trigger')
      expect(steps[2]?.command).toContain(
        `--event-pattern '{"source":["aws.s3"]}'`
      )
      expect(steps[3]?.command).toContain('aws events put-targets')
    })

    test('uses a secret placeholder until the secret is revealed', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'x', preset: 'aws-eventbridge' },
        { ingressUrl: 'https://example.com/ingress' }
      )
      expect(steps[0]?.command).toContain('<SECRET>')
    })

    test('sanitizes hostile names into valid retrigger names', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'x; touch /tmp/pwn', preset: 'aws-eventbridge' },
        { ingressUrl: 'https://example.com/ingress', secret: 's' }
      )
      expect(steps[0]?.command).toContain('--name nuon-x-touch-tmp-pwn')
      expect(steps[0]?.command).not.toContain('--name nuon-x;')
    })

    test('uses the configured topic ARN for SNS subscriptions', () => {
      const steps = buildTriggerSetupSteps(
        {
          name: 'x',
          preset: 'aws-sns',
          auth_config: { topic_arn: 'arn:aws:sns:us-west-2:123:topic' },
        },
        { ingressUrl: 'https://example.com/ingress' }
      )
      expect(steps).toHaveLength(1)
      expect(steps[0]?.command).toContain(
        "--topic-arn 'arn:aws:sns:us-west-2:123:topic'"
      )
      expect(steps[0]?.command).toContain(
        "--notification-endpoint 'https://example.com/ingress'"
      )
    })

    test('omits the Pub/Sub audience flag when no audience is configured', () => {
      const steps = buildTriggerSetupSteps(
        {
          name: 'x',
          preset: 'google-pubsub',
          auth_config: {
            expected_email: 'sa@project.iam.gserviceaccount.com',
          },
        },
        { ingressUrl: 'https://example.com/ingress' }
      )
      expect(steps[0]?.command).toContain(
        "--push-auth-service-account 'sa@project.iam.gserviceaccount.com'"
      )
      expect(steps[0]?.command).not.toContain('--push-auth-token-audience')
    })

    test('includes the first configured Pub/Sub audience', () => {
      const steps = buildTriggerSetupSteps(
        {
          name: 'x',
          preset: 'google-pubsub',
          auth_config: { audience: ['aud-1', 'aud-2'] },
        },
        { ingressUrl: 'https://example.com/ingress' }
      )
      expect(steps[0]?.command).toContain("--push-auth-token-audience 'aud-1'")
      expect(steps[0]?.command).not.toContain('aud-2')
    })

    test('quotes the GitHub webhook secret', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'x', preset: 'github' },
        {
          ingressUrl: 'https://example.com/ingress',
          secret: "s'; rm -rf /; '",
        }
      )
      expect(steps[0]?.command).toContain(
        shellQuote("config[secret]=s'; rm -rf /; '")
      )
      expect(steps[0]?.command).toContain(
        shellQuote('config[url]=https://example.com/ingress')
      )
    })

    test('creates an authenticated Azure Event Grid subscription', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'azure proof', preset: 'azure-event-grid' },
        {
          ingressUrl: 'https://example.com/ingress',
          secret: "s3cr'et",
        }
      )
      expect(steps).toHaveLength(1)
      expect(steps[0]?.command).toContain(
        "--source-resource-id '<SOURCE_RESOURCE_ID>'"
      )
      expect(steps[0]?.command).toContain('--max-events-per-batch 1')
      expect(steps[0]?.command).toContain(
        '--event-delivery-schema eventgridschema'
      )
      expect(steps[0]?.command).toContain(
        `--delivery-attribute-mapping X-Nuon-API-Key static ${shellQuote("s3cr'et")} true`
      )
    })

    test('sanitizes Azure Event Grid subscription names', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'prod_events.v2', preset: 'azure-event-grid' },
        { ingressUrl: 'https://example.com/ingress', secret: 'secret' }
      )
      expect(steps[0]?.command).toContain('--name nuon-prod-events-v2')
    })

    test('builds an Azure Event Grid example payload', () => {
      const curl = buildTriggerCurl(
        {
          preset: 'azure-event-grid',
          envelope: 'none',
          auth_type: 'api_key',
        },
        'https://example.com/ingress',
        'Nuon.Proof.Created'
      )
      expect(curl).toContain('Nuon.Proof.Created')
      expect(curl).toContain('example-event-1')
      expect(curl).toContain('X-Nuon-API-Key: <API_KEY>')
    })

    test('builds Slack Events API setup and signed example request', () => {
      const trigger: TTrigger = {
        name: 'slack proof',
        preset: 'slack-events',
        envelope: 'none',
        auth_type: 'hmac',
        auth_config: {
          header: 'X-Slack-Signature',
          prefix: 'v0=',
          algorithm: 'sha256',
          encoding: 'hex',
        },
      }
      const steps = buildTriggerSetupSteps(trigger, {
        ingressUrl: 'https://example.com/ingress',
      })
      expect(steps).toHaveLength(1)
      expect(steps[0]?.command).toContain('https://example.com/ingress')

      const curl = buildTriggerCurl(
        trigger,
        'https://example.com/ingress',
        'message.channels'
      )
      expect(curl).toContain("TRIGGER_SECRET='<SLACK_SIGNING_SECRET>'")
      expect(curl).toContain('timestamp=$(date +%s)')
      expect(curl).toContain(`printf 'v0:%s:%s' "$timestamp" "$body"`)
      expect(curl).toContain('X-Slack-Request-Timestamp')
      expect(curl).toContain('X-Slack-Signature: v0=')
      expect(curl).toContain('message.channels')
    })

    test('builds Datadog webhook setup with a normalized payload', () => {
      const steps = buildTriggerSetupSteps(
        { name: 'production alerts', preset: 'datadog' },
        {
          ingressUrl: 'https://example.com/ingress',
          secret: "s3cr'et",
        }
      )
      expect(steps).toHaveLength(2)
      expect(steps[0]?.command).toContain(
        '/api/v1/integration/webhooks/configuration/webhooks'
      )
      expect(steps[0]?.command).toContain('X-Nuon-API-Key')
      expect(steps[0]?.command).toContain(`s3cr'"'"'et`)
      expect(steps[0]?.command).toContain('\\"event_id\\":\\"$ID\\"')
      expect(steps[0]?.command).toContain(
        '\\"event_type\\":\\"$EVENT_TYPE\\"'
      )
      expect(steps[1]?.command).toContain('@webhook-nuon-production-alerts')
    })
  })
})
