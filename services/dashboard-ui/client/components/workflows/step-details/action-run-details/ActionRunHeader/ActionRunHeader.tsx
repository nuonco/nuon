import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { IActionRunHeader } from '../types'
import { useInstallLink } from '@/hooks/use-install-path'
import { adhocDisplayName } from '@/utils/action-utils'

interface IActionRunHeaderPresentation extends IActionRunHeader {
  orgId: string
  actionName?: string
}

export const ActionRunHeader = ({
  actionRun,
  isAdhoc,
  loading,
  step,
  orgId,
  actionName,
}: IActionRunHeaderPresentation) => {
  const installLink = useInstallLink()
  const adhocName = adhocDisplayName(
    actionRun?.steps?.at(0)?.adhoc_config?.name
  )
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
          {adhocName ? `Adhoc action: ${adhocName}` : 'Adhoc action'}
        </Text>
      ) : (
        <>
          <Text variant="base" weight="strong">
            {actionName ? `Action: ${actionName}` : 'Action run'}
          </Text>

          <div className="flex items-center gap-4 ml-auto">
            {step?.owner_id && actionRun?.config?.action_workflow_id ? (
              <Link
                href={installLink({
                  orgId: orgId,
                  installId: step.owner_id,
                  suffix: `/actions/${actionRun.config.action_workflow_id}`,
                })}
              >
                View action
              </Link>
            ) : null}
            {step?.owner_id &&
            actionRun?.config?.action_workflow_id &&
            actionRun?.id ? (
              <Link
                href={installLink({
                  orgId: orgId,
                  installId: step.owner_id,
                  suffix: `/actions/${actionRun.config.action_workflow_id}/runs/${actionRun.id}`,
                })}
              >
                View details
              </Link>
            ) : null}
          </div>
        </>
      )}
    </div>
  )
}
