import { useCallback, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstall, TWorkflow } from '@/types'
import { ActiveWorkflowCard } from '../ActiveWorkflowCard'
import { CancelWorkflowsModal } from '../CancelWorkflows'

export const ActiveWorkflows = ({
  workflows,
  install,
}: {
  workflows: TWorkflow[]
  install?: TInstall
}) => {
  const { addModal } = useSurfaces()
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const inProgressWorkflows = workflows.filter(
    (w) => w?.status?.status === 'in-progress'
  )

  const clearSelection = useCallback(() => {
    setSelectedIds(new Set())
  }, [])

  const toggleSelection = useCallback((workflowId: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(workflowId)) {
        next.delete(workflowId)
      } else {
        next.add(workflowId)
      }
      return next
    })
  }, [])

  if (!inProgressWorkflows.length) return null

  const selectionCount = selectedIds.size
  const hasSelection = selectionCount > 0
  const showSelection = inProgressWorkflows.length >= 2

  const openCancelModal = () => {
    addModal(
      <CancelWorkflowsModal
        workflowIds={Array.from(selectedIds)}
        onComplete={clearSelection}
      />
    )
  }

  return (
    <div className="flex flex-col gap-4 @container">
      <div className="flex items-center justify-between gap-2">
        <Text variant="base" weight="strong">
          Active workflows
        </Text>
        {hasSelection && (
          <div className="flex items-center gap-3">
            <Text variant="subtext" theme="info" weight="strong">
              {selectionCount} selected
            </Text>
            <Button variant="secondary" size="sm" onClick={clearSelection}>
              Deselect
            </Button>
            <Button variant="danger" size="sm" onClick={openCancelModal}>
              Cancel workflows
            </Button>
          </div>
        )}
      </div>
      <div className="flex flex-col gap-2">
        {inProgressWorkflows.map((workflow) => (
          <ActiveWorkflowCard
            key={workflow.id}
            workflow={workflow}
            install={install}
            selectable={showSelection}
            selected={selectedIds.has(workflow.id)}
            onToggle={toggleSelection}
            compact={showSelection}
          />
        ))}
      </div>
    </div>
  )
}
