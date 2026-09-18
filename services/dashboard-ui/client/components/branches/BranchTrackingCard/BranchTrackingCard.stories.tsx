export default {
  title: 'Branches/BranchTrackingCard',
}

import { BranchTrackingCard } from './BranchTrackingCard'

export const Default = () => (
  <div className="max-w-3xl">
    <BranchTrackingCard
      repo="acme/kitchen-sink"
      branch="jm/app-branch-tests"
      directory="apps/api"
      latestRun={{
        status: 'success',
        href: '#',
        message: 'chore: exercise github-label-release (#64)',
        author: 'Example Developer',
        sha: '6fdeff3abc123',
        createdAt: '2026-09-17T22:00:00Z',
      }}
    />
  </div>
)

export const WithoutDirectory = () => (
  <div className="max-w-3xl">
    <BranchTrackingCard
      repo="acme/kitchen-sink"
      branch="main"
      latestRun={{
        status: 'active',
        href: '#',
        message: 'feat: update sandbox networking',
        author: 'Example Developer',
        sha: '6987a43',
        createdAt: '2026-09-16T19:58:00Z',
      }}
    />
  </div>
)

export const WithoutCommit = () => (
  <div className="max-w-3xl">
    <BranchTrackingCard
      repo="acme/kitchen-sink"
      branch="main"
      directory="httpbin"
    />
  </div>
)
