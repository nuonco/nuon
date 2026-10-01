import { useMemo, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  computeSummary,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { AppConfigDiffCard } from '@/components/branches/AppConfigDiff/AppConfigDiffCard'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getBranchRunComparison,
  type TBranchRunComparisonConfigDiff,
} from '@/lib'

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

  const summary = sections.length > 0 ? computeSummary(sections) : null

  return (
    <AppConfigDiffCard
      title={title}
      headerAction={headerAction}
      sections={isError ? [] : sections}
      summary={isError ? null : summary}
      isLoading={!isError && isLoading && !data}
      isPending={isPending}
      isOpen
      className={className}
      expandId="branch-overview-config-diff"
    />
  )
}
