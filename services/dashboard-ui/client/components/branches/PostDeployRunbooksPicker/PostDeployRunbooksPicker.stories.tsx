export default {
  title: 'Branches/PostDeployRunbooksPicker',
}

import { useState } from 'react'
import { PostDeployRunbooksPicker } from './PostDeployRunbooksPicker'

const runbooks = [
  { id: 'rb1', name: 'db-migrate' },
  { id: 'rb2', name: 'smoke-test' },
  { id: 'rb3', name: 'warm-cache' },
] as any

const Harness = ({ initial = [] as string[] }) => {
  const [selected, setSelected] = useState<string[]>(initial)
  return (
    <div className="max-w-xl p-4">
      <PostDeployRunbooksPicker
        runbooks={runbooks}
        loadingRunbooks={false}
        selectedRunbookIds={selected}
        onChange={setSelected}
      />
    </div>
  )
}

export const Empty = () => <Harness />

export const WithSelection = () => <Harness initial={['rb1', 'rb2']} />

export const AllSelected = () => <Harness initial={['rb1', 'rb2', 'rb3']} />

export const Loading = () => (
  <div className="max-w-xl p-4">
    <PostDeployRunbooksPicker
      runbooks={[]}
      loadingRunbooks
      selectedRunbookIds={[]}
      onChange={() => {}}
    />
  </div>
)

export const NoRunbooks = () => (
  <div className="max-w-xl p-4">
    <PostDeployRunbooksPicker
      runbooks={[]}
      loadingRunbooks={false}
      selectedRunbookIds={[]}
      onChange={() => {}}
    />
  </div>
)

export const Disabled = () => (
  <div className="max-w-xl p-4">
    <PostDeployRunbooksPicker
      runbooks={runbooks}
      loadingRunbooks={false}
      selectedRunbookIds={['rb1', 'rb2']}
      onChange={() => {}}
      disabled
    />
  </div>
)
