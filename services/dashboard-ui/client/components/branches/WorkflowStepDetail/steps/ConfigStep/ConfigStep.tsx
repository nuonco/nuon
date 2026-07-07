import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

type DiffSummary = { added?: number; removed?: number; changed?: number }

interface IConfigStep {
  appConfigId?: string
  status?: string
  statusDescription?: string
  sections: DiffSectionData[]
  summary: DiffSummary | null
  diffResolved: boolean
  metadata: Record<string, any>
}

export const ConfigStep = ({ appConfigId, status, statusDescription, sections, summary, diffResolved, metadata }: IConfigStep) => {
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

  if (!diffResolved) {
    return <AppConfigDiff sections={[]} summary={null} isLoading />
  }

  if (sections.length === 0) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border">
        <Text variant="subtext" theme="neutral">
          {metadata.component_count !== undefined
            ? `Synced ${metadata.component_count} components${metadata.action_count ? `, ${metadata.action_count} actions` : ''}`
            : 'No changes detected'}
        </Text>
      </div>
    )
  }

  return (
    <AppConfigDiff
      sections={sections}
      summary={summary ? { added: summary.added ?? 0, removed: summary.removed ?? 0, changed: summary.changed ?? 0 } : null}
    />
  )
}
