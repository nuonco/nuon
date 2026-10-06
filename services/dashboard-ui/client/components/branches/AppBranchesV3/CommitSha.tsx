import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import type { TLatestRollout } from './fixtures'

export const CommitSha = ({ commit }: { commit: TLatestRollout['commit'] }) => (
  <Tooltip
    position="top"
    tipContentClassName="whitespace-normal max-w-80"
    tipContent={
      <span className="flex flex-col gap-1.5">
        <Text variant="subtext" className="line-clamp-3 whitespace-normal">
          {commit.message}
        </Text>
        <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <Text variant="subtext" family="mono" theme="neutral">
            {commit.sha.slice(0, 7)}
          </Text>
          <Text variant="subtext" theme="neutral">
            {commit.author}
          </Text>
          <Time
            variant="subtext"
            theme="neutral"
            time={commit.createdAt}
            format="relative"
          />
        </span>
      </span>
    }
  >
    <Text variant="subtext" theme="neutral" family="mono" flex>
      <Icon variant="GitCommitIcon" />
      {commit.sha.slice(0, 7)}
    </Text>
  </Tooltip>
)
