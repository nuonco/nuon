import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { AppConfigDiffCard } from '@/components/branches/AppConfigDiff/AppConfigDiffCard'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

interface IConfigStep {
  appConfigId?: string
  status?: string
  statusDescription?: string
  sections: DiffSectionData[]
  isLoading?: boolean
  isError?: boolean
}

export const ConfigStep = ({
  appConfigId,
  status,
  statusDescription,
  sections,
  isLoading = false,
  isError = false,
}: IConfigStep) => {
  if (!appConfigId) {
    if (status === 'error') {
      return (
        <div className="p-4 bg-red-50 dark:bg-red-950/30 rounded-lg border border-red-200 dark:border-red-800/50">
          <Text variant="subtext" theme="error">
            {statusDescription || 'Configuration sync failed'}
          </Text>
        </div>
      )
    }

    return status === 'in-progress' ? (
      <StepStatePlaceholder variant="loading">
        Cloning repository and parsing configuration
      </StepStatePlaceholder>
    ) : (
      <StepStatePlaceholder variant="pending">
        Waiting to fetch app configuration
      </StepStatePlaceholder>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {status && (
        <div className="flex items-center gap-2">
          <Status status={status} variant="badge" />
          {statusDescription && (
            <Text variant="subtext" theme="neutral">
              {statusDescription}
            </Text>
          )}
        </div>
      )}

      {isError ? (
        <Text variant="subtext" theme="error">
          Unable to load configuration
        </Text>
      ) : (
        <AppConfigDiffCard
          title="Config"
          sections={sections}
          summary={null}
          isLoading={isLoading}
          isOpen
          presentation="snapshot"
          expandId="branch-run-config"
        />
      )}
    </div>
  )
}
