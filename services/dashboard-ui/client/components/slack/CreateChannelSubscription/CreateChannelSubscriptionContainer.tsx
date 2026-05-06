import { useEffect, useMemo, useState } from 'react'
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  createSlackChannelSubscription,
  getSlackChannels,
  getSlackInstallations,
  getSlackOrgLinks,
} from '@/lib'
import type { TAPIError } from '@/types'
import {
  CreateChannelSubscriptionModal,
  type CreateChannelSubscriptionInput,
} from './CreateChannelSubscription'

const CHANNELS_PAGE_SIZE = 100

const CreateChannelSubscriptionModalContainer = (
  props: Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const [selectedInstallationId, setSelectedInstallationId] = useState<
    string | null
  >(null)
  const [channelSearch, setChannelSearch] = useState('')

  const installationsQuery = useQuery({
    queryKey: ['slack-installations', org.id],
    queryFn: () => getSlackInstallations({ orgId: org.id }),
  })

  const orgLinksQuery = useQuery({
    queryKey: ['slack-org-links', org.id],
    queryFn: () => getSlackOrgLinks({ orgId: org.id }),
  })

  const channelsQuery = useInfiniteQuery({
    queryKey: ['slack-channels', org.id, selectedInstallationId],
    queryFn: ({ pageParam }) =>
      getSlackChannels({
        orgId: org.id,
        installationId: selectedInstallationId ?? '',
        types: 'public_channel,private_channel',
        limit: CHANNELS_PAGE_SIZE,
        cursor: pageParam || undefined,
      }),
    initialPageParam: '',
    getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
    enabled: !!selectedInstallationId,
  })

  const channels = useMemo(
    () => channelsQuery.data?.pages.flatMap((p) => p.channels ?? []) ?? [],
    [channelsQuery.data]
  )

  // While the user is searching, eagerly fetch ALL remaining pages so the
  // client-side filter can see every channel in the workspace. Slack's
  // conversations.list has no server-side name filter, so the only way to
  // guarantee complete results is to exhaust pagination.
  useEffect(() => {
    if (!channelSearch.trim()) return
    if (!channelsQuery.hasNextPage) return
    if (channelsQuery.isFetchingNextPage) return
    channelsQuery.fetchNextPage()
  }, [
    channelSearch,
    channelsQuery.data,
    channelsQuery.hasNextPage,
    channelsQuery.isFetchingNextPage,
    channelsQuery.fetchNextPage,
  ])

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: CreateChannelSubscriptionInput) =>
      createSlackChannelSubscription({
        orgId: org.id,
        body: {
          org_link_id: input.orgLinkId,
          channel_id: input.channelId,
          channel_name: input.channelName,
          match: input.match,
          interests: input.interests,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['slack-channel-subscriptions', org.id],
      })
      addToast(
        <Toast heading="Channel subscribed" theme="success">
          <Text>Lifecycle events will start posting here.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      const heading =
        err?.status === 409
          ? 'Channel already subscribed'
          : 'Unable to subscribe channel'
      addToast(
        <Toast heading={heading} theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateChannelSubscriptionModal
      installations={installationsQuery.data ?? []}
      orgLinks={orgLinksQuery.data ?? []}
      channels={channels}
      selectedInstallationId={selectedInstallationId}
      channelsError={channelsQuery.error as TAPIError | null}
      channelSearch={channelSearch}
      onChannelSearchChange={setChannelSearch}
      hasMoreChannels={!!channelsQuery.hasNextPage}
      isLoadingFirstChannelsPage={
        !!selectedInstallationId &&
        channelsQuery.isLoading &&
        channels.length === 0
      }
      isFetchingNextChannelsPage={channelsQuery.isFetchingNextPage}
      onLoadMoreChannels={() => {
        if (channelsQuery.hasNextPage && !channelsQuery.isFetchingNextPage) {
          channelsQuery.fetchNextPage()
        }
      }}
      isPending={isPending}
      error={error}
      onSelectInstallation={(id) => {
        setSelectedInstallationId(id)
        setChannelSearch('')
      }}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const CreateChannelSubscriptionButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateChannelSubscriptionModalContainer />

  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="Plus" />
      Subscribe channel
    </Button>
  )
}
