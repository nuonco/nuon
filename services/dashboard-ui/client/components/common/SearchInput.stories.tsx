import type { Meta, StoryObj } from '@ladle/react'
import { useState } from 'react'
import { SearchInput } from './SearchInput'

export default {
  title: 'Common/SearchInput',
} satisfies Meta

export const Empty: StoryObj = {
  render: () => {
    const [value, setValue] = useState('')
    return (
      <SearchInput
        placeholder="Search..."
        value={value}
        onChange={setValue}
      />
    )
  },
}

export const WithValue: StoryObj = {
  render: () => {
    const [value, setValue] = useState('my search query')
    return (
      <SearchInput
        placeholder="Search..."
        value={value}
        onChange={setValue}
      />
    )
  },
}

export const CustomPlaceholder: StoryObj = {
  render: () => {
    const [value, setValue] = useState('')
    return (
      <SearchInput
        placeholder="Search installs..."
        value={value}
        onChange={setValue}
      />
    )
  },
}
