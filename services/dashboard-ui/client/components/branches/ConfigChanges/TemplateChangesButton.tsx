import { useEffect, useMemo, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TRunSource } from '@/components/branches/BranchOverview/run-source'
import { ConfigChangesLoading } from './ConfigChangesLoading'
import { ConfigChangesViewer } from './ConfigChangesViewer'
import {
  configChanges,
  type TConfigSourceFile,
  type TTemplateBuildChange,
} from './config-changes'

const PANEL_KEY = 'rollout-config-changes'

export interface ITemplateChangesButton {
  sections: DiffSectionData[]
  files?: TConfigSourceFile[]
  builds?: TTemplateBuildChange[]
  versionLabel?: string
  previousSha?: string
  sha?: string
  message?: string
  author?: string
  createdAt?: string
  source?: TRunSource
  isPending?: boolean
  isLoading?: boolean
  isError?: boolean
}

const TemplateChangesPanel = ({
  sections,
  files,
  builds,
  versionLabel,
  previousSha,
  sha,
  message,
  author,
  createdAt,
  source,
  isPending,
  isLoading,
  isError,
  ...props
}: IPanel & ITemplateChangesButton) => (
  <Panel {...props} size="3/4" heading="Changes">
    {isError ? (
      <Text variant="subtext" theme="neutral">
        Changes could not be loaded.
      </Text>
    ) : isPending && sections.length === 0 ? (
      <EmptyState
        emptyTitle="Changes pending"
        emptyMessage="Changes appear after the app config builds."
        variant="diagram"
        size="sm"
      />
    ) : isLoading ? (
      <ConfigChangesLoading />
    ) : (
      <ConfigChangesViewer
        sections={sections}
        files={files}
        builds={builds}
        versionLabel={versionLabel}
        previousSha={previousSha}
        sha={sha}
        message={message}
        author={author}
        createdAt={createdAt}
        source={source}
      />
    )}
  </Panel>
)

export const TemplateChangesButton = (props: ITemplateChangesButton) => {
  const { sections, files, builds, isPending, isLoading } = props
  const { addPanel, updatePanel, panels } = useSurfaces()
  const [searchParams] = useSearchParams()
  const panelId = useRef<string | null>(null)
  const openedByTrigger = useRef(false)
  const openPanelId =
    panels.find((panel) => panel?.id === panelId.current)?.id ?? null
  const { summary } = useMemo(
    () => configChanges(sections, files, builds),
    [builds, files, sections]
  )

  useEffect(() => {
    if (!openPanelId) return
    updatePanel(openPanelId, <TemplateChangesPanel {...props} />)
  }, [
    openPanelId,
    updatePanel,
    sections,
    files,
    builds,
    props.versionLabel,
    props.previousSha,
    props.sha,
    props.message,
    props.author,
    props.createdAt,
    props.source?.kind,
    props.source && 'tag' in props.source ? props.source.tag : undefined,
    props.source && 'number' in props.source ? props.source.number : undefined,
    props.source && 'label' in props.source ? props.source.label : undefined,
    props.isPending,
    props.isLoading,
    props.isError,
  ])

  const open = () => {
    if (openPanelId) return
    panelId.current = addPanel(<TemplateChangesPanel {...props} />, PANEL_KEY)
  }

  const panelParam = searchParams.get('panel')
  useEffect(() => {
    if (openedByTrigger.current) {
      openedByTrigger.current = false
      return
    }
    if (panelParam !== PANEL_KEY || openPanelId) return
    const timer = setTimeout(open, 0)
    return () => clearTimeout(timer)
  }, [panelParam])

  return (
    <Button
      variant="secondary"
      className="gap-3"
      onClick={() => {
        openedByTrigger.current = true
        open()
      }}
    >
      Changes
      {isPending ? (
        <Text variant="subtext" theme="neutral">
          Pending
        </Text>
      ) : isLoading ? null : (
        <ChangeCountSummary
          added={summary.create}
          updated={summary.update}
          removed={summary.delete}
          emptyText="No changes"
        />
      )}
    </Button>
  )
}
