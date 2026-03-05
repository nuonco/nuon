'use client'

import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import type { IActionRunHeader } from './types'

export const ActionRunHeader = ({
  actionRun,
  isAdhoc,
  step,
}: IActionRunHeader) => {
  const { org } = useOrg()

  return (
    <div className="flex items-center gap-4">
      {isAdhoc ? (
        <Text variant="base" weight="strong">
          Adhoc action run
        </Text>
      ) : (
        <>
          <Text variant="base" weight="strong">
            Action run
          </Text>

          <Text variant="subtext">
            <Link
              href={`/${org.id}/installs/${step.owner_id}/actions/${actionRun?.config?.action_workflow_id}`}
            >
              View action <Icon variant="CaretRight" />
            </Link>
          </Text>
          <Text variant="subtext">
            <Link
              href={`/${org.id}/installs/${step.owner_id}/actions/${actionRun?.config?.action_workflow_id}/${actionRun?.id}`}
            >
              View run <Icon variant="CaretRight" />
            </Link>
          </Text>
        </>
      )}
    </div>
  )
}

export const ActionRunHeaderSkeleton = () => {
  return (
    <div className="flex items-center gap-4">
      <Skeleton height="24px" width="76px" />
      <Skeleton height="17" width="85px" />
      <Skeleton height="17" width="70px" />
    </div>
  )
}
