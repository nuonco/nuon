export default {
  title: 'Diffs/DiffSummary',
}

import { DiffSummary } from './DiffSummary'

const summary = {
  create: 4,
  update: 7,
  replace: 1,
  delete: 2,
  read: 0,
  'no-op': 12,
}

export const Default = () => (
  <div className="p-4">
    <DiffSummary summary={summary} />
  </div>
)

export const Subset = () => (
  <div className="p-4">
    <DiffSummary summary={summary} operations={['create', 'update', 'delete']} />
  </div>
)
