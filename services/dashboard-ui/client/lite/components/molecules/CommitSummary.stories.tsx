import { CommitSummary } from './CommitSummary'
import { OverviewCard } from './OverviewCard'

export default {
  title: 'lite/molecules/CommitSummary',
}

export const Default = () => (
  <div className="max-w-sm p-8">
    <OverviewCard title="Last update">
      <CommitSummary
        commit={{
          sha: 'a1b2c3d4e5f6',
          message: 'Update component versions',
          author_name: 'Alex Smith',
          created_at: '2026-09-08T13:42:00Z',
        }}
      />
    </OverviewCard>
  </div>
)

export const Empty = () => (
  <div className="max-w-sm p-8">
    <OverviewCard title="Last update">
      <CommitSummary />
    </OverviewCard>
  </div>
)

export const Loading = () => (
  <div className="max-w-sm p-8">
    <OverviewCard title="Last update">
      <CommitSummary isLoading />
    </OverviewCard>
  </div>
)
