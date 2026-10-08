import { useMemo, useRef, useState } from 'react'
import { cn } from '@/utils/classnames'
import { changeCounts, type TDiffOperation } from '@/lib/diffs'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { DiffEmptyState } from '@/components/diffs/DiffEmptyState'
import { DiffFilter } from '@/components/diffs/DiffFilter'
import { DiffSection } from '@/components/diffs/DiffSection'
import { DiffSections } from '@/components/diffs/DiffSections'
import { DiffSummary } from '@/components/diffs/DiffSummary'
import { usePlanDiffFilter } from '@/components/diffs/use-plan-diff-filter'
import { CommitRange } from '@/components/branches/InstallGroupCards'
import { RunSourceMark } from '@/components/branches/BranchOverview/RunSourceCard'
import type { TRunSource } from '@/components/branches/BranchOverview/run-source'
import { BuildChangeDetail } from './BuildChangeDetail'
import {
  CONFIG_CHANGE_OPERATIONS,
  configChanges,
  type TConfigChangeSection,
  type TConfigSourceFile,
  type TTemplateBuildChange,
} from './config-changes'

const DOT_CLASSES: Record<TDiffOperation, string> = {
  create: 'bg-diff-add',
  update: 'bg-diff-change',
  replace: 'bg-primary-500',
  delete: 'bg-diff-remove',
  read: 'bg-diff-neutral',
  'no-op': 'bg-diff-neutral',
}

const groupSections = (sections: TConfigChangeSection[]) =>
  [...new Set(sections.map(({ group }) => group))].map((group) => {
    const items = sections.filter((section) => section.group === group)
    return { group, icon: items[0].icon, items }
  })

interface IOutlineItem {
  section: TConfigChangeSection
  active: boolean
  onSelect: (id: string) => void
}

const OutlineItem = ({ section, active, onSelect }: IOutlineItem) => {
  const counts = useMemo(
    () => changeCounts(section.before, section.after),
    [section.after, section.before]
  )

  return (
    <button
      type="button"
      onClick={() => onSelect(section.id)}
      className={cn(
        'flex w-full min-w-0 cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-left transition-colors',
        'hover:bg-cool-grey-500/8 active:bg-cool-grey-500/16',
        'focus-visible:[--tw-outline-style:solid] focus-visible:outline-1 focus-visible:outline-offset-0 focus-visible:outline-primary-400/80',
        active && 'bg-cool-grey-500/12'
      )}
    >
      <span
        aria-hidden
        className={cn(
          'size-2 shrink-0 rounded-full',
          DOT_CLASSES[section.operation]
        )}
      />
      <Text
        variant="subtext"
        theme="neutral"
        family={section.kind === 'file' ? 'mono' : undefined}
        className={cn('min-w-0 flex-1 truncate', active && 'text-foreground')}
      >
        {section.title}
      </Text>
      <span className="flex shrink-0 items-center gap-1">
        {counts.added ? (
          <Text variant="label" family="mono" className="text-diff-add">
            +{counts.added}
          </Text>
        ) : null}
        {counts.removed ? (
          <Text variant="label" family="mono" className="text-diff-remove">
            -{counts.removed}
          </Text>
        ) : null}
      </span>
    </button>
  )
}

export interface IConfigChangesViewer {
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
}

export const ConfigChangesViewer = ({
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
}: IConfigChangesViewer) => {
  const changes = useMemo(
    () => configChanges(sections, files, builds),
    [builds, files, sections]
  )
  const filter = usePlanDiffFilter(changes.sections, CONFIG_CHANGE_OPERATIONS)
  const filtered = filter.filteredSections as TConfigChangeSection[]
  const groups = useMemo(() => groupSections(filtered), [filtered])

  const [open, setOpen] = useState<Record<string, boolean>>({})
  const [activeId, setActiveId] = useState<string>()
  const listRef = useRef<HTMLDivElement>(null)

  const setSectionOpen = (id: string, next: boolean) =>
    setOpen((current) =>
      current[id] === next ? current : { ...current, [id]: next }
    )

  const select = (id: string) => {
    setSectionOpen(id, true)
    setActiveId(id)
    const section = listRef.current?.querySelector(
      `[data-config-change="${CSS.escape(id)}"]`
    )
    if (!(section instanceof HTMLElement)) return
    section.scrollIntoView({ behavior: 'smooth', block: 'start' })
    section.querySelector('button')?.focus({ preventScroll: true })
  }

  if (!changes.sections.length) {
    return (
      <div className="rounded-lg bg-cool-grey-100 dark:bg-dark-grey-800 px-4 py-8 text-center">
        <Text as="p" variant="body" weight="strong">
          No changes
        </Text>
        <Text as="p" variant="subtext" theme="neutral">
          This run matches the previous version.
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 md:min-h-0 md:flex-1">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3">
        <span className="flex flex-wrap items-center gap-x-4 gap-y-1">
          {versionLabel ? (
            <Text variant="subtext" theme="neutral" family="mono">
              {versionLabel}
            </Text>
          ) : null}
          {sha || previousSha ? (
            <CommitRange
              commit={{ sha, previousSha, message, author, createdAt }}
            />
          ) : null}
          {source ? <RunSourceMark source={source} /> : null}
        </span>
        <DiffSummary
          summary={changes.summary}
          operations={CONFIG_CHANGE_OPERATIONS}
        />
      </header>

      <div className="grid gap-6 md:min-h-0 md:flex-1 md:grid-cols-[15rem_minmax(0,1fr)]">
        <nav
          aria-label="Changes"
          className="flex flex-col gap-3 md:min-h-0 md:overflow-y-auto md:overscroll-y-contain"
        >
          {groups.map(({ group, icon, items }) => (
            <Expand
              key={group}
              id={`changes-nav-${group.toLowerCase().replace(/\W+/g, '-')}`}
              isOpen
              isIconBeforeHeading
              className="rounded-md"
              headerClassName="rounded-md px-2 py-1.5 text-left"
              heading={
                <span className="flex min-w-0 flex-1 items-center gap-2">
                  <Icon variant={icon} size={16} aria-hidden />
                  <Text
                    variant="subtext"
                    weight="stronger"
                    className="min-w-0 flex-1 truncate"
                  >
                    {group}
                  </Text>
                  <Text variant="label" theme="neutral">
                    {items.length}
                  </Text>
                </span>
              }
            >
              <div className="flex flex-col gap-0.5 pt-1 pl-6">
                {items.map((section) => (
                  <OutlineItem
                    key={section.id}
                    section={section}
                    active={section.id === activeId}
                    onSelect={select}
                  />
                ))}
              </div>
            </Expand>
          ))}
        </nav>

        <div ref={listRef} className="flex min-w-0 flex-col md:min-h-0">
          <DiffSections
            className="md:min-h-0 md:flex-1"
            renderBody={(body) => (
              <Card
                elevation="0"
                className="gap-1 bg-elevation-0 p-0 md:min-h-0 md:flex-1 md:overflow-y-auto md:overscroll-y-contain"
              >
                {body}
              </Card>
            )}
            toolbar={
              <DiffFilter
                title="changes"
                operations={filter.operations}
                selectedOperations={filter.selectedOperations}
                selectedCount={filter.selectedCount}
                totalCount={filter.totalCount}
                searchValue={filter.searchQuery}
                searchPlaceholder="Search changes"
                onSearchChange={filter.setSearchQuery}
                onOperationToggle={filter.toggleOperation}
                onOperationOnly={filter.onlyOperation}
                onReset={filter.reset}
              />
            }
          >
            {groups.length ? (
              groups.flatMap(({ group, icon, items }) => [
                <span
                  key={`group-${group}`}
                  className="flex shrink-0 items-center gap-2 px-1 pt-4 pb-1.5 first:pt-0"
                >
                  <Icon variant={icon} size={16} aria-hidden />
                  <Text as="h4" variant="body" weight="stronger">
                    {group}
                  </Text>
                </span>,
                ...items.map((section) => (
                  <DiffSection
                    key={section.id}
                    data-config-change={section.id}
                    open={open[section.id] ?? true}
                    onOpenChange={(next) => setSectionOpen(section.id, next)}
                    title={section.title}
                    description={section.description}
                    operation={section.operation}
                    before={section.before}
                    after={section.after}
                    language={section.language}
                    filename={section.filename}
                    error={section.error}
                    note={
                      section.build ? (
                        <BuildChangeDetail build={section.build} />
                      ) : undefined
                    }
                    className={cn(
                      'shrink-0 scroll-mt-2',
                      section.id === activeId &&
                        'outline outline-1 outline-primary-400/60'
                    )}
                  />
                )),
              ])
            ) : (
              <DiffEmptyState />
            )}
          </DiffSections>
        </div>
      </div>
    </div>
  )
}
