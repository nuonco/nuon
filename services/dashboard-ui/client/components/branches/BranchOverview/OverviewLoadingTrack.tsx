import { Button } from '@/components/common/Button'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TOverviewStage } from './overview-loading'

const StageLabel = ({ stage }: { stage: TOverviewStage }) => (
  <span className="flex items-center gap-1.5">
    <Status status={stage.status} isWithoutText />
    <Text
      variant="subtext"
      weight={stage.status === 'success' ? 'normal' : 'strong'}
      theme={stage.status === 'pending' ? 'neutral' : undefined}
    >
      {stage.label}
    </Text>
  </span>
)

export const OverviewLoadingTrack = ({
  stages,
  label = 'Run progress',
  activeStageId,
  selectableStageIds,
  onSelectStage,
}: {
  stages: TOverviewStage[]
  label?: string
  activeStageId?: string
  selectableStageIds?: string[]
  onSelectStage?: (stageId: string) => void
}) => (
  <ol
    aria-label={label}
    className="flex min-w-0 flex-wrap items-center gap-y-2"
  >
    {stages.map((stage, index) => (
      <li key={stage.id} className="flex shrink-0 items-center">
        {index > 0 ? (
          <span
            aria-hidden
            className="mx-2.5 h-px w-3 shrink-0 bg-cool-grey-300 dark:bg-white/20"
          />
        ) : null}
        {onSelectStage &&
        (!selectableStageIds || selectableStageIds.includes(stage.id)) ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => onSelectStage(stage.id)}
            className={`!h-auto !gap-1.5 !rounded-md !px-1.5 !py-1 !font-normal ${
              stage.id === activeStageId ? 'bg-cool-grey-500/8' : ''
            }`}
            aria-current={stage.id === activeStageId ? 'step' : undefined}
          >
            <StageLabel stage={stage} />
          </Button>
        ) : (
          <StageLabel stage={stage} />
        )}
      </li>
    ))}
  </ol>
)
