import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
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
  TDatadogEventSubscription,
  TDatadogPriority,
} from '@/types'

export type EditEventSubscriptionInput = {
  match: SubscriptionMatch | undefined
  interests: Interests
  additionalTags: string[]
  alertTypeOverride: TDatadogAlertType | ''
  priorityOverride: TDatadogPriority | ''
}

// Edit modal mirrors the Slack edit-subscription surface: the routing
// identity (connection) is immutable because (connection_id,
// match_canonical) is the unique index — a change there is "delete +
// create", not "update in place". Match, interests, tags, alert type,
// and priority are the editable surfaces.
export const EditEventSubscriptionModal = ({
  subscription,
  connection,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  subscription: TDatadogEventSubscription
  connection: TDatadogConnection | undefined
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: EditEventSubscriptionInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [match, setMatch] = useState<SubscriptionMatch | undefined>(
    subscription.match
  )
  const [interests, setInterests] = useState<Interests>(
    () => subscription.interests ?? allEvents()
  )
  const [tagsText, setTagsText] = useState(
    () => (subscription.additional_tags ?? []).join('\n')
  )
  const [alertOverride, setAlertOverride] = useState<TDatadogAlertType | ''>(
    subscription.alert_type_override ?? ''
  )
  const [priorityOverride, setPriorityOverride] = useState<
    TDatadogPriority | ''
  >(subscription.priority_override ?? '')

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="GraphIcon" size="24" />
          Edit Datadog subscription
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving…
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CheckIcon" />
            Save changes
          </span>
        ),
        disabled: isPending,
        onClick: () =>
          onSubmit({
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
            {error?.description || error?.error || 'Unable to save changes'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label>Datadog connection</Label>
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              {connection?.name || subscription.connection_id || '—'}
            </Text>
            {subscription.connection_id ? (
              <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                {subscription.connection_id}
              </Code>
            ) : null}
          </div>
          <Text variant="subtext" theme="neutral">
            The destination connection is part of the routing identity and
            can't be changed in place. Delete and recreate to point this
            scope at a different connection.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources stream events into this connection.
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
          <Label htmlFor="dd-edit-additional-tags">
            Additional tags{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (one per line, key:value)
            </Text>
          </Label>
          <Textarea
            id="dd-edit-additional-tags"
            placeholder={'customer:acme\nenv:prod'}
            value={tagsText}
            onChange={(e) => setTagsText(e.target.value)}
            rows={2}
          />
          <Text variant="subtext" theme="neutral">
            Appended after the connection's default tags.
          </Text>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-2">
            <Label htmlFor="dd-edit-alert-override">Alert type override</Label>
            <Select
              id="dd-edit-alert-override"
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
            <Label htmlFor="dd-edit-priority-override">
              Priority override
            </Label>
            <Select
              id="dd-edit-priority-override"
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
