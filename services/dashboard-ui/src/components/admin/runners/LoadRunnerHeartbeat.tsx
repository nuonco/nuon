'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Duration } from '@/components/common/Duration'
import { Time } from '@/components/common/Time'
import { LabeledValue } from '@/components/common/LabeledValue'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import type { TRunnerHeartbeat } from '@/types'

interface LoadRunnerHeartbeatProps {
  runnerId: string
}

export const LoadRunnerHeartbeat = ({ runnerId }: LoadRunnerHeartbeatProps) => {
  const { org } = useOrg()
  const orgId = org.id
  
  const { data, error: queryError, isLoading } = useQuery<{ build?: TRunnerHeartbeat; install?: TRunnerHeartbeat }>({
    path: `/api/orgs/${orgId}/runners/${runnerId}/heartbeat`,
    dependencies: [runnerId]
  })

  const heartbeat = data?.build || data?.install
  const error = queryError ? 'Unable to load runner heartbeat' : null

  if (error) {
    return (
      <Text variant="subtext" className="text-red-600">
        {error}
      </Text>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2">
        <Icon variant="Loading" className="animate-spin" size="16" />
        <Text variant="subtext">Loading heartbeat...</Text>
      </div>
    )
  }



  if (!heartbeat?.version) {
    return <Text variant="subtext">No heartbeat yet</Text>
  }

  return (
    <div className="flex items-start gap-4">
      <LabeledValue label="Version">
        {heartbeat.version}
      </LabeledValue>
      <LabeledValue label="Alive time">
        <div className="flex items-center gap-1">
          <Icon variant="Timer"/>
          <Duration variant="subtext" nanoseconds={heartbeat.alive_time} />
        </div>
      </LabeledValue>
      <LabeledValue label="Last heartbeat seen">
        <span className="flex items-center gap-1">
          <Icon variant="Heartbeat" />
          <Time variant="subtext" time={heartbeat.created_at} format="relative" />
        </span>
      </LabeledValue>
    </div>
  )
}
