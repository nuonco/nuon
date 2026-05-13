import { type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface IBranchDetailActions {
  editButton: ReactNode
  manageInstallsButton: ReactNode
  hasConfig: boolean
  isTriggerPending: boolean
  onTriggerRun: () => void
}

export const BranchDetailActions = ({
  editButton,
  manageInstallsButton,
  hasConfig,
  isTriggerPending,
  onTriggerRun,
}: IBranchDetailActions) => {
  return (
    <div className="flex items-center gap-3">
      {editButton}
      {manageInstallsButton}

      <Button
        variant="primary"
        disabled={!hasConfig || isTriggerPending}
        onClick={onTriggerRun}
        title={
          !hasConfig
            ? 'Create a configuration first to trigger a run'
            : 'Trigger a new run with the current configuration'
        }
      >
        <Icon variant="PlayIcon" size={16} />
        {isTriggerPending ? 'Triggering...' : 'Trigger run'}
      </Button>
    </div>
  )
}
