import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TOverviewStage } from './overview-loading'

export const OverviewLoadingTrack = ({
  stages,
}: {
  stages: TOverviewStage[]
}) => (
  <ol
    aria-label="Run progress"
    className="flex flex-wrap items-center gap-x-2 gap-y-3"
  >
    {stages.map((stage, index) => (
      <li key={stage.id} className="flex items-center gap-2">
        {index > 0 ? (
          <Icon
            variant="CaretRightIcon"
            size={12}
            className="text-cool-grey-400"
          />
        ) : null}
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
      </li>
    ))}
  </ol>
)
