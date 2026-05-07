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
  const [cancelMode, setCancelMode] = useState(false)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const inProgressWorkflows = workflows.filter(
    (w) => w?.status?.status === 'in-progress'
  )

  const exitCancelMode = useCallback(() => {
    setCancelMode(false)
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

  const toggleAll = useCallback(() => {
    if (selectedIds.size === inProgressWorkflows.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(inProgressWorkflows.map((w) => w.id)))
    }
  }, [selectedIds.size, inProgressWorkflows])

  if (!inProgressWorkflows.length) return null

  const showBulkCancel = inProgressWorkflows.length >= 2

  return (
    <div className="flex flex-col gap-4 @container">
      <div className="flex items-center justify-between">
        <Text variant="base" weight="strong">
          Active workflows
        </Text>
        {showBulkCancel && !cancelMode && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setCancelMode(true)}
          >
            Cancel workflows
          </Button>
        )}
        {cancelMode && (
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={toggleAll}>
              {selectedIds.size === inProgressWorkflows.length
                ? 'Deselect all'
                : 'Select all'}
            </Button>
            <Button variant="ghost" size="sm" onClick={exitCancelMode}>
              Done
            </Button>
          </div>
        )}
      </div>
      <div className={`grid grid-cols-1 gap-3 ${inProgressWorkflows.length > 1 ? '@3xl:grid-cols-2' : ''}`}>
        {inProgressWorkflows.map((workflow) => (
          <ActiveWorkflowCard
            key={workflow.id}
            workflow={workflow}
            install={install}
            cancelMode={cancelMode}
            selected={selectedIds.has(workflow.id)}
            onToggle={toggleSelection}
          />
        ))}
      </div>
      {cancelMode && selectedIds.size > 0 && (
        <div className="sticky bottom-4 mx-auto w-fit rounded-lg border border-cool-grey-200 dark:border-white/10 bg-code shadow-lg px-4 py-2.5 flex items-center gap-4">
          <Text variant="body">
            {selectedIds.size} workflow{selectedIds.size === 1 ? '' : 's'}{' '}
            selected
          </Text>
          <Button variant="ghost" size="sm" onClick={exitCancelMode}>
            Cancel
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={() => {
              const modal = (
                <CancelWorkflowsModal
                  workflowIds={Array.from(selectedIds)}
                  onComplete={exitCancelMode}
                />
              )
              addModal(modal)
            }}
          >
            Confirm
          </Button>
        </div>
      )}
    </div>
  )
}
