import { CompositeError } from '@/components/common/CompositeError'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { getJobName } from '@/utils/runner-utils'
import type { TRunnerJob } from '@/types'

interface IRunnerJobHeader {
  job: TRunnerJob
}

export const RunnerJobHeader = ({ job }: IRunnerJobHeader) => {
  return (
    <DetailHeader
      title={getJobName(job)}
      id={job?.id}
      identity={
        <Time
          variant="subtext"
          theme="info"
          time={job?.created_at}
          format="relative"
        />
      }
      metadata={
        <>
          <LabeledStatus
            label="Status"
            statusProps={{ status: job?.status_v2?.status ?? job?.status }}
            tooltipProps={{
              tipContent: job?.status_description,
              position: 'bottom',
            }}
          />
          <LabeledValue label="Type">
            <Text variant="subtext">{job?.type}</Text>
          </LabeledValue>
          <LabeledValue label="Group">
            <Text variant="subtext">{job?.group}</Text>
          </LabeledValue>
          <LabeledValue label="Attempts">
            <Text variant="subtext">{job?.execution_count ?? 1}</Text>
          </LabeledValue>
        </>
      }
    >
      {job?.composite_error ? (
        <CompositeError error={job.composite_error} />
      ) : null}
    </DetailHeader>
  )
}
