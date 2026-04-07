import { type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface IBranchDetailActions {
  editButton: ReactNode
  manageInstallsButton: ReactNode
  hasConfig: boolean
  isTriggerPending: boolean
  isCheckPending: boolean
  onTriggerRun: () => void
  onCheckForUpdates: () => void
}

export const BranchDetailActions = ({
  editButton,
  manageInstallsButton,
  hasConfig,
  isTriggerPending,
  isCheckPending,
  onTriggerRun,
  onCheckForUpdates,
}: IBranchDetailActions) => {
  return (
    <div className="flex items-center gap-3">
      {editButton}
      {manageInstallsButton}

      <Button
        variant="secondary"
        disabled={!hasConfig || isCheckPending}
        onClick={onCheckForUpdates}
        title={
          !hasConfig
            ? 'Create a configuration first'
            : 'Check for new commits and trigger a run if updated'
        }
      >
        <Icon variant="ArrowsClockwise" size={16} />
        {isCheckPending ? 'Checking...' : 'Check for updates'}
      </Button>

      <Button
        variant="primary"
        disabled={!hasConfig || isTriggerPending}
        onClick={onTriggerRun}
        title={
          !hasConfig
            ? 'Create a configuration first to trigger a run'
            : 'Force trigger a new run with the current configuration'
        }
      >
        <Icon variant="Play" size={16} />
        {isTriggerPending ? 'Triggering...' : 'Trigger run'}
      </Button>
    </div>
  )
}
