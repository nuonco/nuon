export default {
  title: 'Runners/RunnerCard',
}

import { RunnerCard } from './RunnerCard'

export const Default = () => (
  <RunnerCard
    status="active"
    href="/org-123/installs/inst-456/runner"
  />
)

export const Unhealthy = () => (
  <RunnerCard
    status="unhealthy"
    href="/org-123/installs/inst-456/runner"
  />
)

export const Error = () => (
  <RunnerCard error="No runner found" />
)

export const Loading = () => <RunnerCard isLoading />

export const NoLink = () => (
  <RunnerCard status="active" />
)
