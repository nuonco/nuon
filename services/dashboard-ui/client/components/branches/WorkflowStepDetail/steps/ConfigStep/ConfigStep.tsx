import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

interface IConfigStep {
  appConfigId?: string
  status?: string
  sections: DiffSectionData[]
  isLoading?: boolean
  isError?: boolean
}

const BODY_PADDING = 'px-4 sm:px-6 py-5'

export const ConfigStep = ({
  appConfigId,
  status,
  sections,
  isLoading = false,
  isError = false,
}: IConfigStep) => {
  if (!appConfigId) {
    if (status === 'error') return null

    return (
      <div className={BODY_PADDING}>
        {status === 'in-progress' ? (
          <StepStatePlaceholder variant="loading">
            Cloning repository and parsing configuration
          </StepStatePlaceholder>
        ) : (
          <StepStatePlaceholder variant="pending">
            Waiting to fetch app configuration
          </StepStatePlaceholder>
        )}
      </div>
    )
  }

  if (isError) {
    return (
      <div className={BODY_PADDING}>
        <Text variant="subtext" theme="error">
          Unable to load configuration
        </Text>
      </div>
    )
  }

  return (
    <AppConfigDiff
      sections={sections}
      summary={null}
      isLoading={isLoading}
      defaultSectionsOpen
      presentation="snapshot"
      embedded
    />
  )
}
