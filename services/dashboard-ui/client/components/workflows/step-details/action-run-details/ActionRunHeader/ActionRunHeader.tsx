import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { IActionRunHeader } from '../types'

interface IActionRunHeaderPresentation extends IActionRunHeader {
  orgId: string
}

export const ActionRunHeader = ({
  actionRun,
  isAdhoc,
  step,
  orgId,
}: IActionRunHeaderPresentation) => {
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
              href={`/${orgId}/installs/${step?.owner_id}/actions/${actionRun?.config?.action_workflow_id}`}
            >
              View action <Icon variant="CaretRightIcon" />
            </Link>
          </Text>
          <Text variant="subtext">
            <Link
              href={`/${orgId}/installs/${step?.owner_id}/actions/${actionRun?.config?.action_workflow_id}/runs/${actionRun?.id}`}
            >
              View run details <Icon variant="CaretRightIcon" />
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
