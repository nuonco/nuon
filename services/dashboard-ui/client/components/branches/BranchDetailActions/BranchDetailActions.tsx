import { type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

interface IBranchDetailActions {
  editButton: ReactNode
  manageInstallsButton: ReactNode
  deleteButton: ReactNode
  hasConfig: boolean
  isTriggerPending: boolean
  onTriggerRun: () => void
  onTriggerPreview: () => void
}

export const BranchDetailActions = ({
  editButton,
  manageInstallsButton,
  deleteButton,
  hasConfig,
  isTriggerPending,
  onTriggerRun,
  onTriggerPreview,
}: IBranchDetailActions) => {
  return (
    <div className="flex items-center gap-3">
      {editButton}
      {manageInstallsButton}

      <div className="flex items-center">
        <Button
          variant="primary"
          disabled={!hasConfig || isTriggerPending}
          onClick={onTriggerRun}
          className="!rounded-r-none"
          title={
            !hasConfig
              ? 'Create a configuration first to trigger a run'
              : 'Trigger a new run with the current configuration'
          }
        >
          <Icon variant="PlayIcon" size={16} />
          {isTriggerPending ? 'Triggering...' : 'Trigger run'}
        </Button>

        <Dropdown
          id="trigger-run-options"
          variant="primary"
          alignment="right"
          hideIcon
          disabled={!hasConfig || isTriggerPending}
          buttonClassName="!rounded-l-none !border-l !border-l-primary-700 !px-2"
          buttonText={<Icon variant="CaretDownIcon" size={14} />}
        >
          <Menu>
            <Button isMenuButton onClick={onTriggerPreview}>
              Preview run (plan only)
              <Icon variant="EyeIcon" size={16} />
            </Button>
            <span>{deleteButton}</span>
          </Menu>
        </Dropdown>
      </div>
    </div>
  )
}
