import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

type DiffSummary = { added?: number; removed?: number; changed?: number }

interface IConfigStep {
  appConfigId?: string
  status?: string
  sections: DiffSectionData[]
  summary: DiffSummary | null
  diffResolved: boolean
  metadata: Record<string, any>
}

export const ConfigStep = ({ appConfigId, status, sections, summary, diffResolved, metadata }: IConfigStep) => {
  if (!appConfigId) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Cloning repository and parsing configuration...' : 'Waiting to fetch app configuration...'}
        </Text>
      </div>
    )
  }

  if (!diffResolved) {
    return <AppConfigDiff sections={[]} summary={null} isLoading />
  }

  if (sections.length === 0) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
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
