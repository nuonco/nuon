import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import type { TFieldSize } from '../atoms/Input'
import { SearchInput } from './SearchInput'

export default {
  title: 'lite/molecules/SearchInput',
}

export const Overview = () => (
  <ComponentDocs
    name="SearchInput"
    tier="molecule"
    summary="A controlled search field with consistent search and clear affordances."
    use={[
      'Use for filtering collections, menus, code, and diffs.',
      'Add feature-specific keys through onKeyDown.',
    ]}
    avoid={[
      'Do not rebuild the search icon or clear button at each call site.',
      'Do not use for a value submitted as ordinary form data.',
    ]}
    rules={[
      'Escape clears a non-empty query, and only reaches the surrounding surface once the query is empty. Inside a dropdown that means the first Escape clears and the second closes.',
      'The clear button has an accessible label and appears only when it can act.',
      'The parent controls width and owns what the query searches.',
    ]}
    props={[
      {
        name: 'value',
        type: 'string',
        description: 'Controlled query value.',
      },
      {
        name: 'onValueChange',
        type: '(value: string) => void',
        description: 'Receives typing, clear-button, and Escape changes.',
      },
      {
        name: 'clearLabel',
        type: 'string',
        default: "'Clear search'",
        description: 'Accessible label for the clear button.',
      },
    ]}
  />
)

const SearchExample = ({
  initialValue = '',
  ...props
}: {
  initialValue?: string
  placeholder?: string
  disabled?: boolean
  loading?: boolean
  size?: TFieldSize
}) => {
  const [value, setValue] = useState(initialValue)
  return (
    <SearchInput
      value={value}
      onValueChange={setValue}
      aria-label={props.placeholder ?? 'Search'}
      {...props}
    />
  )
}

export const Default = () => (
  <div className="max-w-md p-8">
    <SearchExample placeholder="Search resources" />
  </div>
)

export const WithValue = () => (
  <div className="max-w-md p-8">
    <SearchExample initialValue="payments" placeholder="Search resources" />
  </div>
)

export const Sizes = () => (
  <div className="flex max-w-md flex-col gap-4 p-8">
    <SearchExample size="sm" placeholder="Search resources" />
    <SearchExample size="md" placeholder="Search resources" />
  </div>
)

export const States = () => (
  <div className="flex max-w-md flex-col gap-4 p-8">
    <SearchExample
      initialValue="payments"
      placeholder="Disabled search"
      disabled
    />
    <SearchExample placeholder="Loading search" loading />
  </div>
)
