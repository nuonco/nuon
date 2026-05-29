import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import {
  InterestsPicker,
  allEvents,
  type Interests,
} from '@/components/interests'
import { MatchPicker } from '@/components/match/MatchPicker'
import type { SubscriptionMatch } from '@/components/match/types'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TDatadogAlertType,
  TDatadogConnection,
  TDatadogPriority,
} from '@/types'

export type CreateEventSubscriptionInput = {
  connectionId: string
  match: SubscriptionMatch | undefined
  interests: Interests
  additionalTags: string[]
  alertTypeOverride: TDatadogAlertType | ''
  priorityOverride: TDatadogPriority | ''
}

export const CreateEventSubscriptionModal = ({
  connections,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  connections: TDatadogConnection[]
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: CreateEventSubscriptionInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  // Default to the first verified connection so a single-connection org
  // doesn't need to interact with the dropdown at all — matches the
  // Slack create flow's "first workspace pre-selected" behavior.
  const firstVerified = connections.find((c) => c.status === 'verified')
  const [connectionId, setConnectionId] = useState(firstVerified?.id ?? '')
  const [match, setMatch] = useState<SubscriptionMatch | undefined>(undefined)
  const [interests, setInterests] = useState<Interests>(() => allEvents())
  const [tagsText, setTagsText] = useState('')
  const [alertOverride, setAlertOverride] = useState<TDatadogAlertType | ''>('')
  const [priorityOverride, setPriorityOverride] = useState<
    TDatadogPriority | ''
  >('')

  const connectionOptions = connections.map((c) => ({
    value: c.id ?? '',
    label: `${c.name ?? '(unnamed)'}${
      c.status === 'revoked' ? ' — revoked' : ''
    }`,
    disabled: c.status !== 'verified',
  }))

  const canSubmit = !!connectionId && !isPending

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="GraphIcon" size="24" />
          Subscribe events to Datadog
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Subscribing…
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Subscribe
          </span>
        ),
        disabled: !canSubmit,
        onClick: () =>
          onSubmit({
            connectionId,
            match,
            interests,
            additionalTags: splitLines(tagsText),
            alertTypeOverride: alertOverride,
            priorityOverride,
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.description || error?.error || 'Unable to create subscription.'}
          </Banner>
        ) : null}

        {connections.length === 0 ? (
          <Banner theme="warn">
            No Datadog connections in this org. Connect a Datadog tenant
            first.
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-connection">Datadog connection</Label>
          <Select
            id="dd-connection"
            options={connectionOptions}
            value={connectionId}
            onChange={(e) => setConnectionId(e.target.value)}
            disabled={connections.length === 0}
            placeholder="Select a connection"
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources stream events into this connection. Leave
            empty to subscribe to every event in the org.
          </Text>
          <MatchPicker value={match} onChange={setMatch} />
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <InterestsPicker
            variant="slack"
            value={interests}
            onChange={setInterests}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-additional-tags">
            Additional tags{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (one per line, key:value)
            </Text>
          </Label>
          <Textarea
            id="dd-additional-tags"
            placeholder={'customer:acme\nenv:prod'}
            value={tagsText}
            onChange={(e) => setTagsText(e.target.value)}
            rows={2}
          />
          <Text variant="subtext" theme="neutral">
            Appended after the connection's default tags. Use to tag events
            with the customer this subscription routes to.
          </Text>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-2">
            <Label htmlFor="dd-alert-override">Alert type override</Label>
            <Select
              id="dd-alert-override"
              options={[
                { value: '', label: 'Default (auto from event)' },
                { value: 'info', label: 'Info' },
                { value: 'warning', label: 'Warning' },
                { value: 'error', label: 'Error' },
                { value: 'success', label: 'Success' },
              ]}
              value={alertOverride}
              onChange={(e) =>
                setAlertOverride(e.target.value as TDatadogAlertType | '')
              }
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="dd-priority-override">Priority override</Label>
            <Select
              id="dd-priority-override"
              options={[
                { value: '', label: 'Default (normal)' },
                { value: 'normal', label: 'Normal' },
                { value: 'low', label: 'Low' },
              ]}
              value={priorityOverride}
              onChange={(e) =>
                setPriorityOverride(e.target.value as TDatadogPriority | '')
              }
            />
          </div>
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
