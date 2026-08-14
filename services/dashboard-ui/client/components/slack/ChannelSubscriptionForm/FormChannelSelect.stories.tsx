import { useForm } from '@tanstack/react-form'
import type { TSlackChannel } from '@/types'
import { FormChannelSelect } from './FormChannelSelect'

export default { title: 'Slack/FormChannelSelect' }

const channels: TSlackChannel[] = [
  { id: 'C0123', name: 'deploys', is_member: true },
  { id: 'C0456', name: 'approvals', is_member: true },
  { id: 'C0789', name: 'general', is_member: true },
]

const Demo = () => {
  const form = useForm({ defaultValues: { channelId: '', channelName: '' } })
  return (
    <div className="max-w-md p-4">
      <form.Field name="channelId">
        {(field) => (
          <FormChannelSelect
            field={field}
            onName={(name) => form.setFieldValue('channelName', name)}
            channels={channels}
            searchQuery=""
            onSearchChange={() => {}}
            onLoadMore={() => {}}
            hasMore={false}
            isLoadingFirstPage={false}
            isFetchingNextPage={false}
            placeholder="Select a channel"
          />
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />
