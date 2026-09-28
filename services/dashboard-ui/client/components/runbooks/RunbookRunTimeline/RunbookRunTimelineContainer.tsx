import type { TInstallRunbookRun } from '@/lib/ctl-api/installs/runbooks/get-install-runbooks'
import { RunbookRunTimeline } from './RunbookRunTimeline'

interface IRunbookRunTimelineContainer {
  runs: TInstallRunbookRun[]
  runbookName: string
  basePath: string
  workflowHref?: (workflowId: string) => string
}

export const RunbookRunTimelineContainer = ({
  runs,
  runbookName,
  basePath,
  workflowHref,
}: IRunbookRunTimelineContainer) => {
  return (
    <RunbookRunTimeline
      runs={runs}
      runbookName={runbookName}
      basePath={basePath}
      workflowHref={workflowHref}
    />
  )
}
