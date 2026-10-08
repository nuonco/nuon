import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { Panel } from '@/components/surfaces/Panel'
import { ConfigChangesViewer } from './ConfigChangesViewer'
import type { TConfigSourceFile } from './config-changes'

export interface ITemplateChangesButton {
  summary?: { added: number; removed: number; changed: number } | null
  sections: DiffSectionData[]
  files?: TConfigSourceFile[]
  versionLabel?: string
  previousSha?: string
  sha?: string
  isPending?: boolean
  isLoading?: boolean
  isError?: boolean
}

export const TemplateChangesButton = ({
  summary,
  sections,
  files,
  versionLabel,
  previousSha,
  sha,
  isPending,
  isLoading,
  isError,
}: ITemplateChangesButton) => (
  <Panel
    panelKey="rollout-config-changes"
    size="3/4"
    heading="Template changes"
    triggerButton={{
      variant: 'secondary',
      className: 'gap-3',
      children: (
        <>
          Template changes
          {isPending ? (
            <Text variant="subtext" theme="neutral">
              Pending
            </Text>
          ) : isLoading ? null : (
            <ChangeCountSummary
              added={summary?.added ?? 0}
              updated={summary?.changed ?? 0}
              removed={summary?.removed ?? 0}
              emptyText="No changes"
            />
          )}
        </>
      ),
    }}
  >
    {isError ? (
      <Text variant="subtext" theme="neutral">
        Template changes could not be loaded.
      </Text>
    ) : isPending && sections.length === 0 ? (
      <EmptyState
        emptyTitle="Changes pending"
        emptyMessage="Changes appear after the app config builds."
        variant="diagram"
        size="sm"
      />
    ) : (
      <ConfigChangesViewer
        sections={sections}
        files={files}
        versionLabel={versionLabel}
        previousSha={previousSha}
        sha={sha}
      />
    )}
  </Panel>
)
