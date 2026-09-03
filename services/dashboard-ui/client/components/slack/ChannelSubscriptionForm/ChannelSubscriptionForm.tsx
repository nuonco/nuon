import { useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { recommendedPreset } from '@/components/interests'
import { FormInterestsPicker } from '@/components/interests/FormInterestsPicker'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TSlackChannel,
  TSlackChannelSubscription,
  TSlackInstallation,
  TSlackOrgLink,
} from '@/types'
import { FormChannelSelect } from './FormChannelSelect'
import {
  buildChannelSubscriptionSchema,
  type ChannelSubscriptionMode,
  type ChannelSubscriptionOutput,
  type ChannelSubscriptionValues,
} from './schema'

export const ChannelSubscriptionFormModal = ({
  mode,
  subscription,
  installations = [],
  orgLinks = [],
  channels = [],
  selectedInstallationId = null,
  channelsError = null,
  channelSearch = '',
  onChannelSearchChange,
  hasMoreChannels = false,
  isLoadingFirstChannelsPage = false,
  isFetchingNextChannelsPage = false,
  onLoadMoreChannels,
  isPending,
  error,
  onSelectInstallation,
  onSubmit,
  ...props
}: {
  mode: ChannelSubscriptionMode
  subscription?: TSlackChannelSubscription
  installations?: TSlackInstallation[]
  orgLinks?: TSlackOrgLink[]
  channels?: TSlackChannel[]
  selectedInstallationId?: string | null
  channelsError?: TAPIError | null
  channelSearch?: string
  onChannelSearchChange?: (q: string) => void
  hasMoreChannels?: boolean
  isLoadingFirstChannelsPage?: boolean
  isFetchingNextChannelsPage?: boolean
  onLoadMoreChannels?: () => void
  isPending: boolean
  error: TAPIError | null
  onSelectInstallation?: (installationId: string) => void
  onSubmit: (output: ChannelSubscriptionOutput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const schema = buildChannelSubscriptionSchema(mode)

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
      channelId: subscription?.channel_id ?? '',
      channelName: subscription?.channel_name ?? '',
      match: subscription?.match,
      interests: subscription?.interests ?? recommendedPreset().build(),
    } as ChannelSubscriptionValues,
    validators: {
      onMount: schema,
      onChange: schema,
    },
    onSubmit: ({ value }) => {
      if (mode === 'create' && !matchingLink?.id) return
      onSubmit({
        orgLinkId:
          mode === 'create'
            ? (matchingLink?.id ?? '')
            : (subscription?.org_link_id ?? ''),
        channelId: value.channelId,
        channelName: value.channelName,
        match: value.match,
        interests: value.interests,
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  useEffect(() => {
    if (mode !== 'create') return
    form.setFieldValue('channelId', '')
    form.setFieldValue('channelName', '')
  }, [mode, selectedInstallationId, form])

  const channelLabel = subscription?.channel_name
    ? `#${subscription.channel_name}`
    : (subscription?.channel_id ?? '—')

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="SlackLogoIcon" size="24" />
          {mode === 'create'
            ? 'Subscribe a channel'
            : 'Edit channel subscription'}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />{' '}
            {mode === 'create' ? 'Subscribing channel' : 'Saving changes'}
          </span>
        ) : mode === 'create' ? (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Subscribe channel
          </span>
        ) : (
          'Save changes'
        ),
        disabled:
          !canSubmit ||
          isPending ||
          (mode === 'create' && !matchingLink?.id),
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
        <FormErrorBanner
          error={error}
          fallback={
            mode === 'create'
              ? 'Unable to subscribe channel'
              : 'Unable to save changes'
          }
        />

        {mode === 'create' ? (
          <>
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
                onChange={(value) => onSelectInstallation?.(value)}
                disabled={installations.length === 0 || isPending}
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
                    disabled={!selectedInstallationId || isPending}
                    placeholder={
                      selectedInstallationId
                        ? 'Select a channel'
                        : 'Pick a workspace first'
                    }
                  />
                )}
              </form.Field>
              <Text variant="subtext" theme="neutral">
                The Nuon bot must be invited to private channels before they
                appear here.
              </Text>
            </div>
          </>
        ) : (
          <>
            <div className="flex flex-col gap-2">
              <Label>Workspace</Label>
              <Text variant="subtext" theme="neutral">
                {subscription?.team_id || '—'}
              </Text>
            </div>

            <div className="flex flex-col gap-2">
              <Label>Channel</Label>
              <div className="flex flex-col gap-1">
                <Text variant="base" weight="strong">
                  {channelLabel}
                </Text>
                {subscription?.channel_id ? (
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
          </>
        )}

        <div className="flex flex-col gap-2">
          <Label>Subscription</Label>
          <Text variant="subtext" theme="neutral">
            Choose a preset to filter which events and which resources post
            notifications in this channel.
          </Text>
          <form.Field name="interests">
            {(field) => (
              <form.Field name="match">
                {(matchField) => (
                  <FormInterestsPicker
                    field={field}
                    matchField={matchField}
                    disabled={isPending}
                  />
                )}
              </form.Field>
            )}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
