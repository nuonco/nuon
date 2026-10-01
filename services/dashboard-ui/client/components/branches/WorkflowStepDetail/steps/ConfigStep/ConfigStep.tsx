import { Text } from '@/components/common/Text'
import { AppConfigFilesDiff } from '@/components/branches/ComponentConfigDiff/ComponentConfigDiff'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { filterExcludedSections } from './lib'
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

  if (isLoading) {
    return (
      <div className={BODY_PADDING}>
        <StepStatePlaceholder variant="loading">
          Loading app configuration
        </StepStatePlaceholder>
      </div>
    )
  }

  // Snapshot of the config as-of this run: no baseline, so no diff badges.
  return (
    <AppConfigFilesDiff
      configSections={filterExcludedSections(sections)}
      files={[]}
      snapshot
      className="border-none shadow-none"
    />
  )
}
