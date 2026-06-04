import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import {
  InterestsPicker,
  allEvents,
  type Interests,
} from '@/components/interests'
import { MatchPicker } from '@/components/match/MatchPicker'
import type { SubscriptionMatch } from '@/components/match/types'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TSlackChannelSubscription } from '@/types'

export type EditChannelSubscriptionInput = {
  match: SubscriptionMatch | undefined
  interests: Interests
}

// Edit-channel-subscription modal. Mirrors the create modal's layout
// (Workspace / Channel / Scope / Events) except identity fields are
// read-only — the underlying row's (team, channel, link) tuple is part of
// the unique index and cannot be edited in place. Match and Interests are
// the only mutable surfaces.
//
// The wire mutation runs PATCH /v1/orgs/{org_id}/slack/channel-subscriptions/
// {sub_id} which uses PUT semantics on `match`: passing `null` (i.e.
// undefined → null over JSON) clears the row to org-wide; passing a
// SubscriptionMatch replaces the existing predicate. A predicate that
// collapses onto an existing row's canonical key returns 409 from the
// backend; the container surfaces that as a "Scope already subscribed to
// this channel" toast.
export const EditChannelSubscriptionModal = ({
  subscription,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  subscription: TSlackChannelSubscription
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: EditChannelSubscriptionInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [match, setMatch] = useState<SubscriptionMatch | undefined>(
    subscription.match
  )
  const [interests, setInterests] = useState<Interests>(
    () => subscription.interests ?? allEvents()
  )

  const channelLabel = subscription.channel_name
    ? `#${subscription.channel_name}`
    : (subscription.channel_id ?? '—')

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="SlackLogoIcon" size="24" />
          Edit channel subscription
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
        onClick: () => onSubmit({ match, interests }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to save changes'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label>Workspace</Label>
          <Text variant="subtext" theme="neutral">
            {subscription.team_id || '—'}
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Channel</Label>
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              {channelLabel}
            </Text>
            {subscription.channel_id ? (
              <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                {subscription.channel_id}
              </Code>
            ) : null}
          </div>
          <Text variant="subtext" theme="neutral">
            Workspace and channel are part of the routing identity and can't
            be changed in place. Delete and recreate to point a different
            channel at this scope.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources fire notifications in this channel.
          </Text>
          <MatchPicker value={match} onChange={setMatch} />
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <Text variant="subtext" theme="neutral">
            Pick which events post notifications in this channel.
          </Text>
          <InterestsPicker value={interests} onChange={setInterests} />
        </div>
      </div>
    </Modal>
  )
}
