import { type ReactNode, useEffect, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'

const NUDGE_DURATION_MS = 8000

interface IBranchDetailActions {
  editButton: ReactNode
  deploymentPlanButton: ReactNode
  deleteButton: ReactNode
  hasDeploymentPlan: boolean
  isTriggerPending: boolean
  showTriggerNudge?: boolean
  onTriggerRun: () => void
  onTriggerPreview: () => void
}

export const BranchDetailActions = ({
  editButton,
  deploymentPlanButton,
  deleteButton,
  hasDeploymentPlan,
  isTriggerPending,
  showTriggerNudge = false,
  onTriggerRun,
  onTriggerPreview,
}: IBranchDetailActions) => {
  const [nudgeOpen, setNudgeOpen] = useState(false)
  const triggerDisabled = !hasDeploymentPlan || isTriggerPending

  useEffect(() => {
    if (!showTriggerNudge) {
      setNudgeOpen(false)
      return
    }
    setNudgeOpen(true)
    const timer = setTimeout(() => setNudgeOpen(false), NUDGE_DURATION_MS)
    return () => clearTimeout(timer)
  }, [showTriggerNudge])

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
        <Tooltip
          isOpen={hasDeploymentPlan ? nudgeOpen : undefined}
          disableHover={hasDeploymentPlan}
          position="bottom"
          tipContent={
            <Text variant="subtext">
              {hasDeploymentPlan
                ? 'Trigger a run to deploy this branch'
                : 'Create a deployment plan to trigger a run'}
            </Text>
          }
        >
          <Button
            variant="primary"
            disabled={triggerDisabled}
            onClick={() => {
              setNudgeOpen(false)
              onTriggerRun()
            }}
            className="!rounded-r-none"
            title="Trigger a new run"
          >
            <Icon variant="PlayIcon" size={16} />
            {isTriggerPending ? 'Triggering...' : 'Trigger run'}
          </Button>
        </Tooltip>

        <Dropdown
          id="trigger-run-options"
          variant="primary"
          alignment="right"
          hideIcon
          disabled={triggerDisabled}
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
