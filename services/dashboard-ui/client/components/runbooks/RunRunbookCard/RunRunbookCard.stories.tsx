export default {
  title: 'Runbooks/RunRunbookCard',
}

import { RunRunbookCard } from './RunRunbookCard'

export const Default = () => (
  <RunRunbookCard
    name="restart-service"
    stepCount={3}
    onRun={() => alert('Run clicked')}
  />
)

export const SingleStep = () => (
  <RunRunbookCard
    name="health-check"
    stepCount={1}
    onRun={() => alert('Run clicked')}
  />
)

export const Loading = () => <RunRunbookCard isLoading />

export const Error = () => (
  <RunRunbookCard error='Runbook "missing-runbook" not found' />
)
