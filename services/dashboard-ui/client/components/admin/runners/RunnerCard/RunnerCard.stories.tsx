export default {
  title: 'Admin/RunnerCard',
}

import { RunnerCard } from './RunnerCard'

const mockRunner = {
  id: 'runner-1',
  status: 'healthy',
  display_name: 'Build Runner',
} as any

export const Default = () => (
  <RunnerCard
    runner={mockRunner}
    href="/org-1/runner"
    isGracefulLoading={false}
    isForceLoading={false}
    isInvalidateLoading={false}
    onGracefulShutdown={() => {}}
    onForceShutdown={() => {}}
    onInvalidateToken={() => {}}
  />
)
