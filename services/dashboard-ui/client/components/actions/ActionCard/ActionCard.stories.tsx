export default {
  title: 'Actions/ActionCard',
}

import { ActionCard } from './ActionCard'

export const WithRun = () => (
  <ActionCard
    name="deploy-canary"
    triggerType="manual"
    status="success"
    href="/org-123/installs/inst-456/actions/action-789"
    hasRun
  />
)

export const CronTrigger = () => (
  <ActionCard
    name="health-check"
    triggerType="cron"
    status="in-progress"
    href="/org-123/installs/inst-456/actions/action-abc"
    hasRun
  />
)

export const FailedRun = () => (
  <ActionCard
    name="rollback"
    triggerType="manual"
    status="error"
    href="/org-123/installs/inst-456/actions/action-def"
    hasRun
  />
)

export const PostDeployTrigger = () => (
  <ActionCard
    name="notify-slack"
    triggerType="post-deploy-component"
    status="success"
    href="/org-123/installs/inst-456/actions/action-jkl"
    hasRun
  />
)

export const NoRuns = () => (
  <ActionCard
    name="cleanup-resources"
    href="/org-123/installs/inst-456/actions/action-ghi"
    hasRun={false}
  />
)

export const Loading = () => <ActionCard isLoading />

export const Error = () => <ActionCard error="Failed to load action" />

export const NotFound = () => (
  <ActionCard error='Action "missing-action" not found' />
)

export const InGroup = () => (
  <div className="flex flex-wrap items-center gap-3">
    <ActionCard
      name="deploy-canary"
      triggerType="manual"
      status="success"
      hasRun
    />
    <ActionCard
      name="health-check"
      triggerType="cron"
      status="in-progress"
      hasRun
    />
    <ActionCard name="pending-action" hasRun={false} />
  </div>
)
