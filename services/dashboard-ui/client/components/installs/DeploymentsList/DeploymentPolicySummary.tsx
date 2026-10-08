import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TDeploymentStep } from '@/components/installs/DeploymentDetail/deployment-progress'

export const DeploymentPolicySummary = ({
  steps,
  history = false,
}: {
  steps: Pick<TDeploymentStep, 'policy'>[]
  history?: boolean
}) => {
  const denyCount = steps.reduce(
    (count, step) => count + (step.policy?.deny_count ?? 0),
    0
  )
  const warnCount = steps.reduce(
    (count, step) => count + (step.policy?.warn_count ?? 0),
    0
  )
  if (!denyCount && !warnCount) return null
  if (history || denyCount > 0) {
    return (
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <Text variant="subtext" theme={denyCount ? 'error' : 'warn'} flex>
          <Icon variant={denyCount ? 'ShieldIcon' : 'WarningIcon'} />
          {denyCount
            ? 'Blocked by policy'
            : `${warnCount} policy warning${warnCount === 1 ? '' : 's'}`}
        </Text>
        {denyCount > 0 ? (
          <>
            <Text variant="subtext" theme="error">
              {denyCount} violation{denyCount === 1 ? '' : 's'}
            </Text>
            {warnCount > 0 ? (
              <Text variant="subtext" theme="warn">
                {warnCount} warning{warnCount === 1 ? '' : 's'}
              </Text>
            ) : null}
          </>
        ) : null}
      </div>
    )
  }
  const firstWarning = steps.find((step) => step.policy?.warn_count)?.policy
    ?.first_warn_message
  return (
    <Banner theme="warn" className="p-3 gap-3">
      <div className="flex flex-col gap-1">
        <Text variant="subtext" weight="strong">
          {warnCount} policy warning{warnCount === 1 ? '' : 's'}
        </Text>
        <div className="flex items-start gap-2">
          <Text
            variant="subtext"
            className="min-w-0 flex-1 !line-clamp-2 break-words"
          >
            {firstWarning || 'Policy warning'}
          </Text>
          {warnCount > 1 ? (
            <Text variant="subtext" className="shrink-0">
              +{warnCount - 1} more
            </Text>
          ) : null}
        </div>
      </div>
    </Banner>
  )
}
