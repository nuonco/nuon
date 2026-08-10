import { useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { allEvents, type Interests } from '@/components/interests'
import { FormInterestsPicker } from '@/components/interests/FormInterestsPicker'
import { FormMatchPicker } from '@/components/match/FormMatchPicker'
import type { SubscriptionMatch } from '@/components/match/types'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TSlackChannel,
  TSlackInstallation,
  TSlackOrgLink,
} from '@/types'
import { FormChannelSelect } from './FormChannelSelect'
import { channelSubscriptionSchema, type ChannelSubscriptionValues } from './schema'

export type CreateChannelSubscriptionInput = {
  orgLinkId: string
  channelId: string
  channelName: string
  match: SubscriptionMatch | undefined
  interests: Interests
}

export const CreateChannelSubscriptionModal = ({
  installations,
  orgLinks,
  channels,
  selectedInstallationId,
  channelsError,
  channelSearch,
  onChannelSearchChange,
  hasMoreChannels,
  isLoadingFirstChannelsPage,
  isFetchingNextChannelsPage,
  onLoadMoreChannels,
  isPending,
  error,
  onSelectInstallation,
  onSubmit,
  ...props
}: {
  installations: TSlackInstallation[]
  orgLinks: TSlackOrgLink[]
  channels: TSlackChannel[]
  selectedInstallationId: string | null
  channelsError: TAPIError | null
  channelSearch: string
  onChannelSearchChange: (q: string) => void
  hasMoreChannels: boolean
  isLoadingFirstChannelsPage: boolean
  isFetchingNextChannelsPage: boolean
  onLoadMoreChannels: () => void
  isPending: boolean
  error: TAPIError | null
  onSelectInstallation: (installationId: string) => void
  onSubmit: (input: CreateChannelSubscriptionInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const installationOptions = useMemo(
    () =>
      installations.map((i) => ({
        value: i.id ?? '',
        label: i.team_name ? `${i.team_name} (${i.team_id})` : (i.team_id ?? ''),
      })),
    [installations]
  )

  const installation = installations.find(
    (i) => i.id === selectedInstallationId
  )
  const matchingLink = orgLinks.find((l) => l.team_id === installation?.team_id)

  const form = useForm({
    defaultValues: {
      channelId: '',
      channelName: '',
      match: undefined,
      interests: allEvents(),
    } as ChannelSubscriptionValues,
    validators: {
      onMount: channelSubscriptionSchema,
      onChange: channelSubscriptionSchema,
    },
    onSubmit: ({ value }) => {
      if (!matchingLink?.id) return
      onSubmit({
        orgLinkId: matchingLink.id,
        channelId: value.channelId,
        channelName: value.channelName,
        match: value.match,
        interests: value.interests,
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  useEffect(() => {
    form.setFieldValue('channelId', '')
    form.setFieldValue('channelName', '')
  }, [selectedInstallationId, form])

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="SlackLogoIcon" size="24" />
          Subscribe a channel
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
            Subscribe channel
          </span>
        ),
        disabled: !canSubmit || !matchingLink?.id || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner error={error} fallback="Unable to subscribe channel" />

        {installations.length === 0 ? (
          <Banner theme="warn">
            No Slack workspaces are connected to this org. Install the Nuon
            Slack app first.
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="slack-installation">Workspace</Label>
          <Select
            id="slack-installation"
            options={installationOptions}
            value={selectedInstallationId ?? ''}
            placeholder="Select a workspace"
            onChange={(value) => onSelectInstallation(value)}
            disabled={installations.length === 0}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="slack-channel">Channel</Label>
          {channelsError ? (
            <Banner theme="error">
              {channelsError?.error ||
                'Unable to load channels for this workspace.'}
            </Banner>
          ) : null}
          <form.Field name="channelId">
            {(field) => (
              <FormChannelSelect
                field={field}
                onName={(name) => form.setFieldValue('channelName', name)}
                id="slack-channel"
                channels={channels}
                searchQuery={channelSearch}
                onSearchChange={onChannelSearchChange}
                onLoadMore={onLoadMoreChannels}
                hasMore={hasMoreChannels}
                isLoadingFirstPage={isLoadingFirstChannelsPage}
                isFetchingNextPage={isFetchingNextChannelsPage}
                disabled={!selectedInstallationId}
                placeholder={
                  selectedInstallationId
                    ? 'Select a channel'
                    : 'Pick a workspace first'
                }
              />
            )}
          </form.Field>
          <Text variant="subtext" theme="neutral">
            The Nuon bot must be invited to private channels before they appear
            here.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources fire notifications in this channel.
          </Text>
          <form.Field name="match">
            {(field) => <FormMatchPicker field={field} />}
          </form.Field>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <Text variant="subtext" theme="neutral">
            Pick which events post notifications in this channel.
          </Text>
          <form.Field name="interests">
            {(field) => <FormInterestsPicker field={field} />}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
