import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Text } from '../atoms/Text'
import { ListSearch } from './ListSearch'

export default {
  title: 'lite/molecules/ListSearch',
}

export const Overview = () => (
  <ComponentDocs
    name="ListSearch"
    tier="molecule"
    summary="A debounced SearchInput for URL-backed collection queries."
    use={[
      'Use with useListQueryState when a collection searches through an API.',
      'Keep SearchInput for immediate client-side filtering.',
    ]}
    avoid={[
      'Do not put fetching or query-key logic in this component.',
      'Do not use it for ordinary form fields.',
    ]}
    rules={[
      'The visible draft updates immediately while commits are debounced.',
      'A controlled value change from Back or Forward replaces the visible draft.',
      'Unmounting cancels a pending commit.',
    ]}
    props={[
      {
        name: 'value',
        type: 'string',
        description: 'Committed URL-backed search value.',
      },
      {
        name: 'onValueChange',
        type: '(value: string) => void',
        description: 'Receives the debounced search value.',
      },
      {
        name: 'debounceMs',
        type: 'number',
        default: '300',
        description: 'Delay between the final keystroke and commit.',
      },
    ]}
  />
)

const SearchExample = ({
  initialValue = '',
  debounceMs,
}: {
  initialValue?: string
  debounceMs?: number
}) => {
  const [value, setValue] = useState(initialValue)

  return (
    <div className="flex max-w-md flex-col gap-3 p-8">
      <ListSearch
        value={value}
        onValueChange={setValue}
        debounceMs={debounceMs}
        placeholder="Search installs"
        aria-label="Search installs"
      />
      <Text variant="caption" color="secondary">
        Committed query: {value || 'None'}
      </Text>
    </div>
  )
}

export const Default = () => <SearchExample />

export const WithValue = () => <SearchExample initialValue="payments" />

export const CustomDebounce = () => <SearchExample debounceMs={1000} />
