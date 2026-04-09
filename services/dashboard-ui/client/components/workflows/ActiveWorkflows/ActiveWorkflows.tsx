import { Text } from '@/components/common/Text'
import type { TInstall, TWorkflow } from '@/types'
import { ActiveWorkflowCard } from '../ActiveWorkflowCard'

export const ActiveWorkflows = ({
  workflows,
  install,
}: {
  workflows: TWorkflow[]
  install?: TInstall
}) => {
  const inProgressWorkflows = workflows.filter(
    (w) => w?.status?.status === 'in-progress'
  )

  if (!inProgressWorkflows.length) return null

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Active workflows
      </Text>
      <div className="flex flex-col gap-3">
        {inProgressWorkflows.map((workflow) => (
          <ActiveWorkflowCard
            key={workflow.id}
            workflow={workflow}
            install={install}
          />
        ))}
      </div>
    </div>
  )
}
