import { useEffect, useState } from 'react'
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
  TDatadogConnection,
  TDatadogConnectionPurpose,
  TUpdateDatadogConnectionBody,
} from '@/types'

export const EditConnectionModal = ({
  connection,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  connection: TDatadogConnection
  isPending: boolean
  error: TAPIError | null
  onSubmit: (body: TUpdateDatadogConnectionBody) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState(connection.name ?? '')
  const [site, setSite] = useState(connection.site ?? 'us1')
  const [apiKey, setApiKey] = useState('')
  const [appKey, setAppKey] = useState('')
  const [purpose, setPurpose] = useState<TDatadogConnectionPurpose>(
    (connection.purpose as TDatadogConnectionPurpose) ?? 'internal'
  )
  const [tagsText, setTagsText] = useState(
    (connection.default_tags ?? []).join('\n')
  )
  const [handlesText, setHandlesText] = useState(
    (connection.default_notify_handles ?? []).join('\n')
  )

  useEffect(() => {
    setName(connection.name ?? '')
    setSite(connection.site ?? 'us1')
    setPurpose(
      (connection.purpose as TDatadogConnectionPurpose) ?? 'internal'
    )
    setTagsText((connection.default_tags ?? []).join('\n'))
    setHandlesText((connection.default_notify_handles ?? []).join('\n'))
  }, [connection])

  // PATCH semantics: only send fields the user actually changed. The
  // backend re-validates the API key only when it's present, so leaving
  // the key field blank is the explicit "don't rotate" signal.
  const buildPatch = (): TUpdateDatadogConnectionBody => {
    const patch: TUpdateDatadogConnectionBody = {}
    if (name.trim() !== (connection.name ?? '')) patch.name = name.trim()
    if (site.trim() !== (connection.site ?? '')) patch.site = site.trim()
    if (purpose !== ((connection.purpose as TDatadogConnectionPurpose) ?? 'internal')) {
      patch.purpose = purpose
    }
    if (apiKey.trim()) patch.api_key = apiKey.trim()
    if (appKey.trim()) patch.application_key = appKey.trim()

    const tags = splitLines(tagsText)
    if (!arraysEqual(tags, connection.default_tags ?? [])) {
      patch.default_tags = tags
    }
    const handles = splitLines(handlesText)
    if (!arraysEqual(handles, connection.default_notify_handles ?? [])) {
      patch.default_notify_handles = handles
    }
    return patch
  }

  const canSubmit = !isPending

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="GraphIcon" size="24" />
          Edit Datadog connection
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving…
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit,
        onClick: () => onSubmit(buildPatch()),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.description || error?.error || 'Unable to update connection.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-name">Display name</Label>
          <Input
            id="dd-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
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
        </div>

        <SiteInput value={site} onChange={setSite} />

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-api-key">
            Rotate API key{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (leave blank to keep current)
            </Text>
          </Label>
          <Input
            id="dd-api-key"
            type="password"
            placeholder="New DD API key"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            autoComplete="off"
          />
          <Text variant="subtext" theme="neutral">
            New keys are re-validated against Datadog before saving.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-app-key">
            Rotate application key{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (leave blank to keep current)
            </Text>
          </Label>
          <Input
            id="dd-app-key"
            type="password"
            placeholder="New DD application key"
            value={appKey}
            onChange={(e) => setAppKey(e.target.value)}
            autoComplete="off"
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-default-tags">Default tags</Label>
          <Textarea
            id="dd-default-tags"
            value={tagsText}
            onChange={(e) => setTagsText(e.target.value)}
            rows={3}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-default-handles">Default notification handles</Label>
          <Textarea
            id="dd-default-handles"
            value={handlesText}
            onChange={(e) => setHandlesText(e.target.value)}
            rows={3}
          />
        </div>
      </div>
    </Modal>
  )
}

const splitLines = (s: string): string[] =>
  s
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0)

const arraysEqual = (a: string[], b: string[]): boolean => {
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false
  return true
}
