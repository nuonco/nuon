import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import { stepStatusCategory } from '@/components/branches/shared/step-status'

export type TDeploymentStage = {
  id: string
  name: string
  status: string
}

export const DeploymentStageRail = ({
  stages,
  currentId,
  onSelect,
}: {
  stages: TDeploymentStage[]
  currentId?: string
  onSelect: (stageId: string) => void
}) => (
  <ol className="flex min-w-0 flex-1 gap-1" aria-label="Deployment plan">
    {stages.map((stage) => {
      const current = stage.id === currentId
      const category = stepStatusCategory(stage.status)
      return (
        <li key={stage.id} className="flex min-w-0 flex-1">
          <Tooltip
            position="top"
            className="block"
            style={{ width: '100%' }}
            tipContent={
              <span className="flex items-center gap-2">
                <Text variant="subtext" weight="strong">
                  {stage.name}
                </Text>
                <Status status={stage.status} />
              </span>
            }
          >
            <button
              type="button"
              className="flex h-6 w-full cursor-pointer items-center rounded-sm outline-none focus-visible:shadow-[0_0_0_1px_var(--background),0_0_0_3px_var(--focus-ring)]"
              aria-current={current ? 'step' : undefined}
              aria-label={`${stage.name}, ${stage.status}`}
              onClick={() => onSelect(stage.id)}
            >
              <span
                className={cn(
                  'w-full rounded-full',
                  current ? 'h-2' : 'h-1',
                  category === 'success' &&
                    'bg-green-600 dark:bg-green-500',
                  (category === 'active' || category === 'awaiting') &&
                    'bg-blue-600 dark:bg-blue-500',
                  category === 'error' && 'bg-red-600 dark:bg-red-500',
                  category === 'pending' &&
                    'bg-cool-grey-200 dark:bg-dark-grey-700'
                )}
              />
            </button>
          </Tooltip>
        </li>
      )
    })}
  </ol>
)
