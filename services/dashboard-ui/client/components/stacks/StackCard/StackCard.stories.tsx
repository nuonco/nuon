export default {
  title: 'Stacks/StackCard',
}

import { StackCard } from './StackCard'

export const Default = () => (
  <StackCard
    status="active"
    runCount={5}
    createdAt="2025-01-15T10:30:00Z"
    href="/org-123/installs/inst-456/stacks"
  />
)

export const NoRuns = () => (
  <StackCard
    status="active"
    runCount={0}
    createdAt="2025-01-15T10:30:00Z"
    href="/org-123/installs/inst-456/stacks"
  />
)

export const SingleRun = () => (
  <StackCard
    status="active"
    runCount={1}
    createdAt="2025-01-15T10:30:00Z"
    href="/org-123/installs/inst-456/stacks"
  />
)

export const Provisioning = () => (
  <StackCard
    status="provisioning"
    runCount={2}
    createdAt="2025-03-01T08:00:00Z"
    href="/org-123/installs/inst-456/stacks"
  />
)

export const Error = () => <StackCard error="Failed to load stack" />

export const NoStack = () => <StackCard error="No stack found" />

export const Loading = () => <StackCard isLoading />

export const NoLink = () => (
  <StackCard status="active" runCount={3} createdAt="2025-02-20T14:00:00Z" />
)
