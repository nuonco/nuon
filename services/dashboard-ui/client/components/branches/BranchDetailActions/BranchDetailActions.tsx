import { type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

interface IBranchDetailActions {
  editButton: ReactNode
  deploymentPlanButton: ReactNode
  deleteButton: ReactNode
  hasConfig: boolean
  isTriggerPending: boolean
  onTriggerRun: () => void
  onTriggerPreview: () => void
}

export const BranchDetailActions = ({
  editButton,
  deploymentPlanButton,
  deleteButton,
  hasConfig,
  isTriggerPending,
  onTriggerRun,
  onTriggerPreview,
}: IBranchDetailActions) => {
  return (
    <div className="flex items-center gap-3">
      <Dropdown
        id="branch-manage"
        variant="secondary"
        alignment="right"
        buttonText={
          <>
            <Icon variant="SlidersHorizontalIcon" size={16} />
            Manage
          </>
        }
      >
        <Menu className="min-w-56">
          {deploymentPlanButton}
          {editButton}
          <hr />
          <span className="contents">{deleteButton}</span>
        </Menu>
      </Dropdown>

      <div className="flex items-center">
        <Button
          variant="primary"
          disabled={!hasConfig || isTriggerPending}
          onClick={onTriggerRun}
          className="!rounded-r-none"
          title={
            !hasConfig
              ? 'Create a deployment plan first to trigger a run'
              : 'Trigger a new run with the current deployment plan'
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
          </Menu>
        </Dropdown>
      </div>
    </div>
  )
}
