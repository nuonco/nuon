import { useState } from 'react'
import type { TPermissionVerb } from '@/types'
import { ALL_VERBS } from '../permissions'
import { VerbsDropdown } from './VerbsDropdown'

export default {
  title: 'Access roles/VerbsDropdown',
}

const Harness = ({ initial }: { initial: TPermissionVerb[] }) => {
  const [value, setValue] = useState(initial)

  return <VerbsDropdown id="story-verbs" value={value} onChange={setValue} />
}

export const AllActions = () => <Harness initial={[...ALL_VERBS]} />

export const Subset = () => <Harness initial={['read', 'update']} />

export const None = () => <Harness initial={[]} />

export const Disabled = () => (
  <VerbsDropdown
    id="story-verbs-disabled"
    value={[...ALL_VERBS]}
    onChange={() => {}}
    disabled
  />
)
