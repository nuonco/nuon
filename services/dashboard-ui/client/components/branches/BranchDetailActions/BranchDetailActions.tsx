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
  isTriggerPending: boolean
  showManage?: boolean
  showTriggerNudge?: boolean
  onTriggerRun: () => void
  onTriggerPreview: () => void
}

export const BranchDetailActions = ({
  editButton,
  deploymentPlanButton,
  deleteButton,
  isTriggerPending,
  showManage = true,
  showTriggerNudge = false,
  onTriggerRun,
  onTriggerPreview,
}: IBranchDetailActions) => {
  const [nudgeOpen, setNudgeOpen] = useState(false)

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
      {showManage ? (
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
      ) : null}

      <div className="flex items-center">
        <Tooltip
          isOpen={nudgeOpen}
          disableHover
          position="bottom"
          tipContent={
            <Text variant="subtext">Trigger a run to deploy this branch</Text>
          }
        >
          <Button
            variant="primary"
            disabled={isTriggerPending}
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
          disabled={isTriggerPending}
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
