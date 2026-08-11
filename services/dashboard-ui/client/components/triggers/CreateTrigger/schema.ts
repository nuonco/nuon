import { z } from 'zod'

export type TProvider =
  | 'github'
  | 'slack'
  | 'datadog'
  | 'aws'
  | 'gcp'
  | 'azure'
  | 'custom'
export type TAwsMethod = 'aws-eventbridge' | 'aws-sns'

export const getTriggerPreset = (provider: TProvider, awsMethod: TAwsMethod) =>
  provider === 'aws'
    ? awsMethod
    : provider === 'gcp'
      ? 'google-pubsub'
      : provider === 'azure'
        ? 'azure-event-grid'
        : provider === 'slack'
          ? 'slack-events'
          : provider

export const createTriggerSchema = z
  .object({
    name: z.string().trim().min(1, 'Name is required'),
    description: z.string(),
    provider: z.enum([
      'github',
      'slack',
      'datadog',
      'aws',
      'gcp',
      'azure',
      'custom',
    ]),
    awsMethod: z.enum(['aws-eventbridge', 'aws-sns']),
    authType: z.enum([
      'none',
      'hmac',
      'api_key',
      'basic',
      'bearer_jwt',
      'sns_signature',
    ]),
    envelope: z.enum(['none', 'pubsub_push', 'cloudevents', 'sns']),
    issuer: z.string(),
    audience: z.string(),
    identity: z.string(),
    topicArn: z.string(),
    header: z.string(),
    username: z.string(),
    slackSigningSecret: z.string(),
  })
  .superRefine((v, ctx) => {
    const preset = getTriggerPreset(v.provider, v.awsMethod)
    const custom = v.provider === 'custom'

    if (preset === 'google-pubsub' || (custom && v.authType === 'bearer_jwt')) {
      if (!v.issuer.trim())
        ctx.addIssue({ code: 'custom', path: ['issuer'], message: 'Issuer is required' })
      if (!v.audience.trim())
        ctx.addIssue({ code: 'custom', path: ['audience'], message: 'Audience is required' })
      if (!v.identity.trim())
        ctx.addIssue({ code: 'custom', path: ['identity'], message: 'Expected identity is required' })
    }

    if (preset === 'aws-sns' || (custom && v.authType === 'sns_signature')) {
      if (!v.topicArn.trim())
        ctx.addIssue({ code: 'custom', path: ['topicArn'], message: 'SNS topic ARN is required' })
    }

    if (preset === 'slack-events' && !v.slackSigningSecret.trim()) {
      ctx.addIssue({
        code: 'custom',
        path: ['slackSigningSecret'],
        message: 'Slack signing secret is required',
      })
    }
  })

export type CreateTriggerValues = z.infer<typeof createTriggerSchema>
