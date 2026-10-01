import { useMemo, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import {
  computeSummary,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/diffs/plan-diff-switch'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getBranchRunComparison,
  type TBranchRunComparisonConfigDiff,
} from '@/lib'
import { cn } from '@/utils/classnames'

const GROUPED_SECTIONS = new Set([
  'Components',
  'Actions',
  'Runbooks',
  'Install inputs',
  'Secrets',
  'Policies',
])

export function summarySectionsFromComparisonConfigDiff(
  content?: TBranchRunComparisonConfigDiff | null
): DiffSectionData[] {
  if (!content?.sections?.length) return []

  return content.sections.map((sec) => {
    const grouped = GROUPED_SECTIONS.has(sec.name)
    const entities = grouped
      ? sec.entries.map((e) => ({
          name: e.name,
          op: (e.op as 'add' | 'remove' | 'change') || 'change',
          fields: e.description
            ? [{ key: 'change', op: e.op, diff: e.description }]
            : e.source_changed
              ? [{ key: 'source', op: 'change', diff: 'source files changed' }]
              : [
                  {
                    key: 'change',
                    op: e.op || 'change',
                    diff: 'Configuration changed',
                  },
                ],
        }))
      : []

    return {
      name: sec.name,
      sectionKey: sec.name.toLowerCase().replace(/\s+/g, '_'),
      additions: sec.additions,
      removals: sec.removals,
      changed: sec.changed,
      grouped,
      entities,
      fields: !grouped
        ? sec.entries.flatMap((e) =>
            e.description
              ? [{ key: e.name, op: e.op, diff: e.description }]
              : []
          )
        : [],
    }
  })
}

interface IBranchRunChangesSummary {
  branchId: string
  appBranchRunId: string
  className?: string
  title?: string
  headerAction?: ReactNode
  isPending?: boolean
}

export const BranchRunChangesSummary = ({
  branchId,
  appBranchRunId,
  className,
  title = 'Config Changes',
  headerAction,
  isPending,
}: IBranchRunChangesSummary) => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data, isLoading, isError } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-run-comparison',
      org?.id,
      app?.id,
      branchId,
      appBranchRunId,
      'config',
    ],
    queryFn: () =>
      getBranchRunComparison({
        orgId: org!.id,
        appId: app!.id,
        branchId,
        runId: appBranchRunId,
        includeDiff: ['config'],
      }),
    enabled: !!org?.id && !!app?.id && !!branchId && !!appBranchRunId,
    retry: 1,
  })

  const sections = useMemo(
    () => summarySectionsFromComparisonConfigDiff(data?.config_diff_content),
    [data?.config_diff_content]
  )

  const visibleSections = isError ? [] : sections
  const summary =
    visibleSections.length > 0 ? computeSummary(visibleSections) : null
  const loading = !isError && isLoading && !data
  const showPending = isPending && visibleSections.length === 0

  return (
    <section
      className={cn(
        'border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden min-w-0',
        className
      )}
    >
      <header className="flex items-center justify-between gap-3 px-5 py-4">
        <Text variant="h3" weight="strong">
          {title}
        </Text>
        <div className="flex items-center gap-3">
          {showPending ? (
            <Text variant="subtext" theme="neutral">
              Pending
            </Text>
          ) : !loading ? (
            <ChangeCountSummary
              added={summary?.added ?? 0}
              updated={summary?.changed ?? 0}
              removed={summary?.removed ?? 0}
              emptyText="No changes"
            />
          ) : null}
          {headerAction}
        </div>
      </header>
      <div className="border-t max-h-[70vh] overflow-y-auto">
        {showPending ? (
          <div className="px-4 py-6 text-center">
            <EmptyState
              emptyTitle="Changes pending"
              emptyMessage="Changes appear after the app config builds."
              variant="diagram"
              size="sm"
            />
          </div>
        ) : (
          <AppConfigDiff
            sections={visibleSections}
            summary={null}
            isLoading={loading}
            defaultSectionsOpen={false}
            embedded
          />
        )}
      </div>
    </section>
  )
}
