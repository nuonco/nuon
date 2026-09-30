import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TOverviewStage } from './overview-loading'

const StageLabel = ({ stage }: { stage: TOverviewStage }) => (
  <>
    <Status
      status={stage.status}
      variant="timeline"
      isWithoutText
      iconSize={14}
    />
    <Text
      variant="subtext"
      theme={stage.status === 'pending' ? 'neutral' : undefined}
    >
      {stage.label}
    </Text>
  </>
)

export const OverviewLoadingTrack = ({
  stages,
  label = 'Run progress',
  onSelectStage,
}: {
  stages: TOverviewStage[]
  label?: string
  onSelectStage?: (stageId: string) => void
}) => (
  <ol aria-label={label} className="flex flex-wrap items-center gap-x-2 gap-y-3">
    {stages.map((stage, index) => (
      <li key={stage.id} className="flex items-center gap-2">
        {index > 0 ? (
          <Icon
            variant="CaretRightIcon"
            size={12}
            className="text-cool-grey-400"
          />
        ) : null}
        {onSelectStage ? (
          <button
            type="button"
            onClick={() => onSelectStage(stage.id)}
            className="flex items-center gap-2 rounded-md px-1 py-0.5 hover:bg-black/[0.03] dark:hover:bg-white/[0.03]"
          >
            <StageLabel stage={stage} />
          </button>
        ) : (
          <StageLabel stage={stage} />
        )}
      </li>
    ))}
  </ol>
)
