import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  computeSummary,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import { AppConfigDiffCard } from '@/components/branches/AppConfigDiff/AppConfigDiffCard'
import { BranchRunComparisonRuns } from '@/components/branches/BranchRunComparisonRuns'
import { Card } from '@/components/common/Card'
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

export function sectionsFromComparisonConfigDiff(
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
              : [],
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

interface IBranchRunChanges {
  branchId: string
  appBranchRunId: string
  focus?: TConfigDiffFocus | null
  className?: string
  showRunComparison?: boolean
  repoSlug?: string
}

export const BranchRunChanges = ({
  branchId,
  appBranchRunId,
  focus,
  className,
  showRunComparison = true,
  repoSlug,
}: IBranchRunChanges) => {
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
    () => sectionsFromComparisonConfigDiff(data?.config_diff_content),
    [data?.config_diff_content]
  )

  const summary = sections.length > 0 ? computeSummary(sections) : null

  const showComparison =
    showRunComparison &&
    !!org?.id &&
    !!app?.id &&
    (data?.head_run || data?.base_run)

  if (isError) {
    return (
      <AppConfigDiffCard
        title="Config Changes"
        sections={[]}
        summary={null}
        isLoading={false}
        isOpen
        className={className}
      />
    )
  }

  return (
    <div className={`flex flex-col gap-4 ${className ?? ''}`}>
      {showComparison ? (
        <Card className="!p-4 !gap-3">
          <BranchRunComparisonRuns
            orgId={org!.id}
            appId={app!.id}
            branchId={branchId}
            baseRun={data?.base_run}
            headRun={data?.head_run}
            repoSlug={repoSlug}
          />
        </Card>
      ) : null}

      <AppConfigDiffCard
        title="Config Changes"
        sections={sections}
        summary={summary}
        isLoading={isLoading && !data}
        isOpen
        focus={focus}
        expandId="branch-run-config-diff"
      />
    </div>
  )
}
