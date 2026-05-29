import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import { SiteInput } from '@/components/datadog/SiteInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TCreateDatadogConnectionBody,
  TDatadogConnectionPurpose,
} from '@/types'

export type CreateConnectionInput = TCreateDatadogConnectionBody

export const CreateConnectionModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: CreateConnectionInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState('')
  const [site, setSite] = useState('us1')
  const [apiKey, setApiKey] = useState('')
  const [appKey, setAppKey] = useState('')
  const [purpose, setPurpose] = useState<TDatadogConnectionPurpose>('internal')
  // tagsText keeps the user's literal newline-delimited input so we can
  // re-render their edits verbatim. Backend wants string[] — split on
  // submit only.
  const [tagsText, setTagsText] = useState('')
  const [handlesText, setHandlesText] = useState('')

  const canSubmit =
    name.trim().length > 0 &&
    site.trim().length > 0 &&
    apiKey.trim().length >= 10 &&
    !isPending

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="GraphIcon" size="24" />
          Connect Datadog
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Connecting…
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Connect
          </span>
        ),
        disabled: !canSubmit,
        onClick: () => {
          if (!canSubmit) return
          onSubmit({
            name: name.trim(),
            site: site.trim(),
            api_key: apiKey.trim(),
            application_key: appKey.trim() || undefined,
            purpose,
            default_tags: splitLines(tagsText),
            default_notify_handles: splitLines(handlesText),
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.description || error?.error || 'Unable to create connection.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-name">Display name</Label>
          <Input
            id="dd-name"
            placeholder="Internal monitoring"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <Text variant="subtext" theme="neutral">
            Shown in the dashboard list. Customers see their own connection's
            name, not yours.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-purpose">Purpose</Label>
          <Select
            id="dd-purpose"
            options={[
              { value: 'internal', label: 'Internal — your team' },
              { value: 'customer', label: 'Customer — vendored to a customer DD' },
            ]}
            value={purpose}
            onChange={(e) =>
              setPurpose(e.target.value as TDatadogConnectionPurpose)
            }
          />
          <Text variant="subtext" theme="neutral">
            Cosmetic badge in the connection list. Doesn't affect routing.
          </Text>
        </div>

        <SiteInput value={site} onChange={setSite} />

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-api-key">API key</Label>
          <Input
            id="dd-api-key"
            type="password"
            placeholder="DD API key"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            autoComplete="off"
          />
          <Text variant="subtext" theme="neutral">
            Verified against Datadog before saving. Stored plaintext on the
            backend; treat the key as a shared secret you can rotate.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-app-key">
            Application key{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (optional)
            </Text>
          </Label>
          <Input
            id="dd-app-key"
            type="password"
            placeholder="DD application key"
            value={appKey}
            onChange={(e) => setAppKey(e.target.value)}
            autoComplete="off"
          />
          <Text variant="subtext" theme="neutral">
            Only needed for the one-click "Alert in Datadog" monitor action.
            Event-stream delivery works without it.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-default-tags">
            Default tags{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (one per line, key:value)
            </Text>
          </Label>
          <Textarea
            id="dd-default-tags"
            placeholder={'env:prod\nteam:platform'}
            value={tagsText}
            onChange={(e) => setTagsText(e.target.value)}
            rows={3}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-default-handles">
            Default notification handles{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (one per line, @-prefixed)
            </Text>
          </Label>
          <Textarea
            id="dd-default-handles"
            placeholder={'@pagerduty-prod\n@slack-oncall'}
            value={handlesText}
            onChange={(e) => setHandlesText(e.target.value)}
            rows={3}
          />
          <Text variant="subtext" theme="neutral">
            Spliced into one-click monitor bodies so DD's @-mention fan-out
            routes to your alerting tools. Override per-monitor at create
            time.
          </Text>
        </div>
      </div>
    </Modal>
  )
}

// splitLines turns the textarea contents into a clean string[]. Empty
// lines and surrounding whitespace get stripped so a stray trailing
// newline doesn't render a "  " ghost tag in DD.
const splitLines = (s: string): string[] =>
  s
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0)
