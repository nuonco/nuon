import { useMemo, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import {
  computeSummary,
  extractSections,
  type DiffChangeKind,
  type DiffEntityEntry,
  type DiffFieldEntry,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TBuildMeta } from '@/components/branches/BranchOverview/changed-builds'
import {
  scopedComparisonConfigDiff,
  type TComparisonScope,
} from '@/components/branches/BranchRunChanges/comparison-scope'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/diffs/plan-diff-switch'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppConfigDiff,
  getBranchRunComparison,
  type TBranchRunComparisonConfigDiff,
  type TBranchRunComparisonConfigDiffEntry,
} from '@/lib'
import type { TCompositeError } from '@/types'
import { cn } from '@/utils/classnames'
import {
  configDiagnosticLines,
  isConfigValidationError,
} from './config-diagnostics'
import { ConfigParseFailure } from './ConfigParseFailure'

const GROUPED_SECTIONS = new Set([
  'Components',
  'Actions',
  'Runbooks',
  'Install inputs',
  'Secrets',
  'Policies',
  'Permissions',
])

const SECTION_KEYS: Record<string, string> = {
  Components: 'components',
  Actions: 'actions',
  Runbooks: 'runbooks',
  'Install inputs': 'inputs',
  Secrets: 'secrets',
  Policies: 'policies',
  Sandbox: 'sandbox',
  Runner: 'runner',
  Permissions: 'permissions',
  Stack: 'stack',
  'Break glass': 'break_glass',
  'Operation roles': 'operation_roles',
}

const entryField = (
  entry: TBranchRunComparisonConfigDiffEntry
): DiffFieldEntry => {
  if (entry.description) {
    return { key: entry.name, op: entry.op, diff: entry.description }
  }
  if (entry.file) {
    return { key: 'file', op: entry.op, diff: entry.file }
  }
  return {
    key: 'change',
    op: entry.op || 'change',
    diff: 'Configuration changed',
  }
}

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
      sectionKey:
        SECTION_KEYS[sec.name] ??
        sec.name.toLowerCase().replace(/\s+/g, '_'),
      additions: sec.additions,
      removals: sec.removals,
      changed: sec.changed,
      grouped,
      entities,
      fields: !grouped ? sec.entries.map(entryField) : [],
    }
  })
}

const kindsForReason = (reason?: string): DiffChangeKind[] => {
  if (reason === 'source_changed') return ['source']
  if (reason === 'config_changed') return ['config']
  if (reason === 'source_and_config') return ['source', 'config']
  return []
}

const orderedKinds = (kinds: Set<DiffChangeKind>): DiffChangeKind[] => {
  const out: DiffChangeKind[] = []
  if (kinds.has('source')) out.push('source')
  if (kinds.has('config')) out.push('config')
  return out
}

const isSourceOnlyEntity = (entity: DiffEntityEntry) =>
  entity.fields.length > 0 && entity.fields.every((field) => field.key === 'source')

const isSandboxBuild = (build: TBuildMeta) =>
  build.component_type === 'sandbox' || build.component_id === 'sandbox'

// Config-diff rows stay. Components the build step rebuilt for source, and that
// the config diff omitted, are added and marked Source, Config, or both.
export function withBuildChangeKinds(
  sections: DiffSectionData[],
  builds: TBuildMeta[] = []
): DiffSectionData[] {
  const byName = new Map<string, Set<DiffChangeKind>>()
  for (const build of builds) {
    if (isSandboxBuild(build)) continue
    const kinds = kindsForReason(build.change_reason)
    if (kinds.length === 0) continue
    const name = build.component_name || build.component_id
    if (!name) continue
    const set = byName.get(name) ?? new Set<DiffChangeKind>()
    for (const kind of kinds) set.add(kind)
    byName.set(name, set)
  }

  const next = sections.map((section) => ({
    ...section,
    entities: section.entities.map((entity) => ({ ...entity })),
  }))

  let components = next.find((section) => section.sectionKey === 'components')
  if (!components && byName.size > 0) {
    components = {
      name: 'Components',
      sectionKey: 'components',
      additions: 0,
      removals: 0,
      changed: 0,
      grouped: true,
      entities: [],
      fields: [],
    }
    next.unshift(components)
  }
  if (!components) return next

  const seen = new Set<string>()
  components.entities = components.entities.map((entity) => {
    seen.add(entity.name)
    const kinds = new Set<DiffChangeKind>()
    if (isSourceOnlyEntity(entity)) kinds.add('source')
    else kinds.add('config')
    for (const kind of byName.get(entity.name) ?? []) kinds.add(kind)
    return { ...entity, changeKinds: orderedKinds(kinds) }
  })

  for (const [name, kinds] of byName) {
    if (seen.has(name)) continue
    components.entities.push({
      name,
      op: 'change',
      changeKinds: orderedKinds(kinds),
      fields: [],
    })
    components.changed += 1
  }

  return next
}

export function overlaySectionDetail(
  summary: DiffSectionData[],
  detailed: DiffSectionData[]
): DiffSectionData[] {
  if (!detailed.length) return summary
  const byKey = new Map(detailed.map((section) => [section.sectionKey, section]))
  return summary.map((section) => {
    if (section.grouped) return section
    const detail = byKey.get(section.sectionKey)
    if (!detail?.content && !detail?.fields.length && !detail?.files?.length) {
      return section
    }
    return {
      ...section,
      fields: detail.fields,
      files: detail.files,
      content: detail.content,
    }
  })
}

interface IBranchRunChangesSummary {
  branchId: string
  appBranchRunId: string
  builds?: TBuildMeta[]
  className?: string
  title?: string
  headerAction?: ReactNode
  isPending?: boolean
  scope?: TComparisonScope
  configError?: TCompositeError
}

export const BranchRunChangesSummary = ({
  branchId,
  appBranchRunId,
  builds = [],
  className,
  title = 'Config Changes',
  headerAction,
  isPending,
  scope,
  configError,
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

  const headConfigId = data?.head_run?.app_config_id
  const baseConfigId = data?.base_run?.app_config_id
  const { data: configDiff } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config-diff', org?.id, app?.id, headConfigId, baseConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: headConfigId!,
        oldConfigId: baseConfigId,
      }),
    enabled: !!org?.id && !!app?.id && !!headConfigId,
    retry: 1,
  })

  const sections = useMemo(() => {
    const summary = summarySectionsFromComparisonConfigDiff(
      scopedComparisonConfigDiff(data?.config_diff_content, scope)
    )
    const detailed = configDiff?.diff ? extractSections(configDiff.diff) : []
    return withBuildChangeKinds(overlaySectionDetail(summary, detailed), builds)
  }, [data?.config_diff_content, configDiff?.diff, builds, scope])

  const visibleSections = isError ? [] : sections
  const summary =
    visibleSections.length > 0 ? computeSummary(visibleSections) : null
  const loading = !isError && isLoading && !data
  const showPending = isPending && visibleSections.length === 0
  const showConfigError =
    isConfigValidationError(configError) &&
    !showPending &&
    visibleSections.length === 0

  if (showConfigError) {
    return (
      <ConfigParseFailure
        className={className}
        title={title}
        headerAction={headerAction}
        lines={configDiagnosticLines(configError)}
      />
    )
  }

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
