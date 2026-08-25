import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { IActionRunHeader } from '../types'

interface IActionRunHeaderPresentation extends IActionRunHeader {
  orgId: string
}

export const ActionRunHeader = ({
  actionRun,
  isAdhoc,
  loading,
  step,
  orgId,
}: IActionRunHeaderPresentation) => {
  if (loading) {
    return (
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Action run
        </Text>
        <Text variant="subtext" loading loadingWidth={12} />
        <Text variant="subtext" loading loadingWidth={12} />
      </div>
    )
  }

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

          {step?.owner_id && actionRun?.config?.action_workflow_id ? (
            <Link
              href={`/${orgId}/installs/${step.owner_id}/actions/${actionRun.config.action_workflow_id}`}
            >
              View action
            </Link>
          ) : null}
          {step?.owner_id &&
          actionRun?.config?.action_workflow_id &&
          actionRun?.id ? (
            <Link
              href={`/${orgId}/installs/${step.owner_id}/actions/${actionRun.config.action_workflow_id}/runs/${actionRun.id}`}
            >
              View run details
            </Link>
          ) : null}
        </>
      )}
    </div>
  )
}
