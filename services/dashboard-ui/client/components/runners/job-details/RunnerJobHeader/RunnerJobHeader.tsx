import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { getJobName } from '@/utils/runner-utils'
import type { TRunnerJob } from '@/types'

interface IRunnerJobHeader {
  job: TRunnerJob
}

export const RunnerJobHeader = ({ job }: IRunnerJobHeader) => {
  return (
    <header className="p-6 border-b flex justify-between">
      <HeadingGroup>
        <BackLink className="mb-6" />
        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            {getJobName(job)}
          </Text>
          <ID>{job.id}</ID>
          <Time variant="subtext" time={job.created_at} />
        </div>
      </HeadingGroup>

      <div className="flex gap-6 items-start">
        <LabeledValue label="Status">
          <Status status={job.status} />
        </LabeledValue>
        <LabeledValue label="Type">
          <Text variant="subtext">{job.type}</Text>
        </LabeledValue>
        <LabeledValue label="Group">
          <Text variant="subtext">{job.group}</Text>
        </LabeledValue>
      </div>
    </header>
  )
}
