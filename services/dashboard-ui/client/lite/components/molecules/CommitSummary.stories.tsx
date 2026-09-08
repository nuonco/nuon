import { ComponentDocs } from '../__stories__/ComponentDocs'
import { CommitSummary } from './CommitSummary'
import { OverviewCard } from './OverviewCard'

export default {
  title: 'lite/molecules/CommitSummary',
}

export const Overview = () => (
  <ComponentDocs
    name="CommitSummary"
    tier="molecule"
    summary="A short commit identity: abbreviated SHA, subject line, author, and relative time."
    use={[
      'Describe the commit behind a branch run or an install config sync.',
      'Compose inside an OverviewCard titled for the update it describes.',
    ]}
    avoid={[
      'Do not use it to render a full commit history or a list of commits.',
      'Do not assume a commit has a message, an author, or a timestamp.',
    ]}
    rules={[
      'The SHA is abbreviated to seven characters.',
      'Only the first line of the commit message is shown.',
      'The empty state explains that the latest synced commit will appear later.',
      'It renders fragments, so the parent owns layout and spacing.',
    ]}
    props={[
      {
        name: 'commit',
        type: 'TVCSCommit',
        description: 'Commit to summarize. Renders an empty state when absent.',
      },
      {
        name: 'updatedAt',
        type: 'string',
        description:
          'Fallback timestamp used when the commit carries no created date.',
      },
      {
        name: 'isLoading',
        type: 'boolean',
        default: 'false',
        description: 'Shows SHA, message, and time loading shapes.',
      },
    ]}
  />
)

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
