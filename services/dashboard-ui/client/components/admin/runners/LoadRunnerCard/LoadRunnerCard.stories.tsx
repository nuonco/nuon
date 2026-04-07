export default {
  title: 'Admin/LoadRunnerCard',
}

import { LoadRunnerCard } from './LoadRunnerCard'

const mockRunner = {
  id: 'runner-1',
  status: 'healthy',
  display_name: 'Test Runner',
} as any

export const Default = () => (
  <LoadRunnerCard
    runner={mockRunner}
    error={null}
    isLoading={false}
    href="/org-1/installs/install-1/runner"
    onAction={() => {}}
  />
)

export const Loading = () => (
  <LoadRunnerCard
    runner={undefined}
    error={null}
    isLoading={true}
    href="/org-1/installs/install-1/runner"
    onAction={() => {}}
  />
)

export const Error = () => (
  <LoadRunnerCard
    runner={undefined}
    error="Unable to load runner"
    isLoading={false}
    href="/org-1/installs/install-1/runner"
    onAction={() => {}}
  />
)

export const NotFound = () => (
  <LoadRunnerCard
    runner={undefined}
    error={null}
    isLoading={false}
    href="/org-1/installs/install-1/runner"
    onAction={() => {}}
  />
)
