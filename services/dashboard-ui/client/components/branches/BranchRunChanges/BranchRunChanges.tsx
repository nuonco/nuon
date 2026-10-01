import { useMemo, useState, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import {
  AppConfigFilesDiff,
  type TComponentSourceFile,
} from '@/components/branches/ComponentConfigDiff/ComponentConfigDiff'
import { BranchRunComparisonRuns } from '@/components/branches/BranchRunComparisonRuns'
import { Card } from '@/components/common/Card'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppConfigSourceFile,
  getBranchRunComparison,
  type TSourceArchiveFileDiff,
  type TBranchRunComparisonConfigDiff,
} from '@/lib'

const GROUPED_SECTIONS = new Set([
  'Components',
  'Actions',
  'Runbooks',
  'Policies',
  'Permissions',
])

// Sections agreed out of scope for the run comparison view: their diffs stay
// visible through the full config diff, not here.
const EXCLUDED_SECTIONS = new Set(['Stack', 'Install inputs', 'Secrets'])

export function sectionsFromComparisonConfigDiff(
  content?: TBranchRunComparisonConfigDiff | null
): DiffSectionData[] {
  if (!content?.sections?.length) return []

  return content.sections
    .filter((sec) => !EXCLUDED_SECTIONS.has(sec.name))
    .map((sec) => {
      const grouped = GROUPED_SECTIONS.has(sec.name)
      const entityOp = (op: string): 'add' | 'remove' | 'change' =>
        op === 'add' ? 'add' : op === 'remove' ? 'remove' : 'change'

      // Line diffs live only in the file tree; entity rows carry the entity's
      // defining file so clicking it focuses that file in the tree.
      const entities = grouped
        ? sec.entries.map((e) => ({
            name: e.name,
            op: entityOp(e.op),
            fields: [],
            files: e.file ? [{ name: e.file, op: entityOp(e.op) }] : undefined,
          }))
        : []

      const first = sec.entries[0]
      const sectionFile =
        !grouped && first?.file
          ? [{ name: first.file, op: entityOp(first.op) }]
          : undefined

      return {
        name: sec.name,
        sectionKey: sec.name.toLowerCase().replace(/\s+/g, '_'),
        additions: sec.additions,
        removals: sec.removals,
        changed: sec.changed,
        grouped,
        entities,
        fields: [],
        files: sectionFile,
      }
    })
}

const sourceFileChange = (op: string) =>
  op === 'added'
    ? 'added'
    : op === 'removed'
      ? 'removed'
      : op === 'modified'
        ? 'modified'
        : 'unchanged'

const sourceFilesFromDiff = (
  files: TSourceArchiveFileDiff[] | undefined
): TComponentSourceFile[] =>
  (files ?? []).map((file) => ({
    path: file.path,
    kind: 'file',
    change: sourceFileChange(file.op),
  }))

interface IBranchRunChanges {
  branchId: string
  appBranchRunId: string
  className?: string
  showRunComparison?: boolean
  repoSlug?: string
  title?: string
  headerAction?: ReactNode
  isPending?: boolean
}

export const BranchRunChanges = ({
  branchId,
  appBranchRunId,
  className,
  showRunComparison = true,
  repoSlug,
  title,
  headerAction,
  isPending,
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
        includeDiff: ['config', 'source'],
      }),
    enabled: !!org?.id && !!app?.id && !!branchId && !!appBranchRunId,
    retry: 1,
  })

  const sections = useMemo(
    () => sectionsFromComparisonConfigDiff(data?.config_diff_content),
    [data?.config_diff_content]
  )

  const sourceFiles = useMemo(
    () => sourceFilesFromDiff(data?.source_diff_content?.files),
    [data?.source_diff_content?.files]
  )
  const [selectedPath, setSelectedPath] = useState<string | undefined>()
  const selectedFile = data?.source_diff_content?.files.find(
    ({ path }) => path === selectedPath
  )
  const headConfigId = data?.head_run?.app_config_id
  const baseConfigId = data?.base_run?.app_config_id
  const needBefore =
    !!selectedFile &&
    (selectedFile.op === 'removed' || selectedFile.op === 'modified') &&
    !!baseConfigId
  const needAfter =
    !!selectedFile && selectedFile.op !== 'removed' && !!headConfigId

  const beforeQuery = useQuery({
    queryKey: [
      'branch-run-source-file-before',
      org?.id,
      app?.id,
      baseConfigId,
      selectedPath,
    ],
    queryFn: () =>
      getAppConfigSourceFile({
        orgId: org!.id,
        appId: app!.id,
        configId: baseConfigId!,
        path: selectedPath!,
      }),
    enabled: needBefore,
  })
  const afterQuery = useQuery({
    queryKey: [
      'branch-run-source-file-after',
      org?.id,
      app?.id,
      headConfigId,
      selectedPath,
    ],
    queryFn: () =>
      getAppConfigSourceFile({
        orgId: org!.id,
        appId: app!.id,
        configId: headConfigId!,
        path: selectedPath!,
      }),
    enabled: needAfter,
  })
  const loadingFile =
    (needBefore && beforeQuery.isPending) || (needAfter && afterQuery.isPending)

  const filesWithContents = useMemo(
    () =>
      sourceFiles.map((file) =>
        file.path !== selectedPath
          ? file
          : {
              ...file,
              before:
                needBefore && !beforeQuery.isPending
                  ? beforeQuery.data?.content
                  : undefined,
              after:
                needAfter && !afterQuery.isPending
                  ? afterQuery.data?.content
                  : undefined,
            }
      ),
    [
      sourceFiles,
      selectedPath,
      needBefore,
      needAfter,
      beforeQuery.data,
      beforeQuery.isPending,
      afterQuery.data,
      afterQuery.isPending,
    ]
  )

  const showComparison =
    showRunComparison &&
    !!org?.id &&
    !!app?.id &&
    (data?.head_run || data?.base_run)

  if (isError) {
    return (
      <AppConfigFilesDiff
        title={title}
        headerAction={headerAction}
        isPending={isPending}
        previousVersion={data?.base_sha ?? ''}
        currentVersion={data?.head_sha ?? ''}
        configSections={[]}
        files={[]}
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

      <AppConfigFilesDiff
        title={title}
        headerAction={headerAction}
        isPending={isPending}
        previousVersion={data?.base_sha ?? ''}
        currentVersion={data?.head_sha ?? ''}
        configSections={sections}
        files={filesWithContents}
        onSelectPath={setSelectedPath}
        isLoadingFile={loadingFile}
      />
    </div>
  )
}
