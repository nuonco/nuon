export default { title: 'Features / Operations / Actor filter' }

import { useState } from 'react'
import { ActorFilter } from './ActorFilter'

const CURRENT_USER = { accountId: 'acc-me', email: 'jane@example.com' }

const MEMBERS = [
  { accountId: 'acc-1', email: 'joe@example.com' },
  { accountId: 'acc-2', email: 'priya@example.com' },
  { accountId: 'acc-3', email: 'marcus@example.com' },
]

export const Default = () => {
  const [selected, setSelected] = useState<string>()
  const [query, setQuery] = useState('')
  return (
    <div className="flex justify-end p-6 pb-96 max-w-xl">
      <ActorFilter
        id="actor-filter-story"
        currentUser={CURRENT_USER}
        members={MEMBERS}
        query={query}
        onQueryChange={setQuery}
        selected={selected}
        onChange={setSelected}
      />
    </div>
  )
}

export const SelectedTeammate = () => {
  const [selected, setSelected] = useState<string | undefined>('acc-1')
  const [query, setQuery] = useState('')
  return (
    <div className="flex justify-end p-6 pb-96 max-w-xl">
      <ActorFilter
        id="actor-filter-story-selected"
        currentUser={CURRENT_USER}
        members={MEMBERS}
        query={query}
        onQueryChange={setQuery}
        selected={selected}
        onChange={setSelected}
      />
    </div>
  )
}

export const SelectedMe = () => {
  const [selected, setSelected] = useState<string | undefined>(
    CURRENT_USER.accountId
  )
  const [query, setQuery] = useState('')
  return (
    <div className="flex justify-end p-6 pb-96 max-w-xl">
      <ActorFilter
        id="actor-filter-story-me"
        currentUser={CURRENT_USER}
        members={MEMBERS}
        query={query}
        onQueryChange={setQuery}
        selected={selected}
        onChange={setSelected}
      />
    </div>
  )
}

export const Loading = () => {
  const [query, setQuery] = useState('')
  return (
    <div className="flex justify-end p-6 pb-96 max-w-xl">
      <ActorFilter
        id="actor-filter-story-loading"
        currentUser={CURRENT_USER}
        members={[]}
        isLoading
        query={query}
        onQueryChange={setQuery}
        onChange={() => {}}
      />
    </div>
  )
}

export const NoMatches = () => {
  const [query, setQuery] = useState('zzz')
  return (
    <div className="flex justify-end p-6 pb-96 max-w-xl">
      <ActorFilter
        id="actor-filter-story-empty"
        currentUser={CURRENT_USER}
        members={[]}
        query={query}
        onQueryChange={setQuery}
        onChange={() => {}}
      />
    </div>
  )
}
