import { useEffect, useMemo, useState } from 'react'
import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { Menu } from '@/components/common/Menu'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { Select } from '@/components/common/form/Select'
import { cn } from '@/utils/classnames'
import { downloadFileOnClick } from '@/utils/file-download'
import { BlockEditor } from './BlockEditor'
import {
  compileTemplate,
  getArraySources,
  getStateVariables,
  previewBlock,
  previewDocument,
} from './compiler'
import { seedBlocks } from './seed'
import type { TBlock, TBlockType, TEntityOption } from './types'

const blockMeta: Record<
  TBlockType,
  { label: string; icon: TIconVariant; hint: string }
> = {
  markdown: {
    label: 'Markdown',
    icon: 'TextAaIcon',
    hint: 'Freeform markdown with {{.nuon.*}} variables',
  },
  banner: {
    label: 'Banner',
    icon: 'WarningIcon',
    hint: 'Callout, optionally shown on a condition',
  },
  'status-row': {
    label: 'Status row',
    icon: 'PulseIcon',
    hint: 'Inline badges and values from install state',
  },
  table: {
    label: 'Table',
    icon: 'TableIcon',
    hint: 'Repeats rows over a list in install state',
  },
  runbook: {
    label: 'Run runbook',
    icon: 'BookOpenIcon',
    hint: 'Card that runs an app runbook from the install page',
  },
  action: {
    label: 'Run action',
    icon: 'PlayIcon',
    hint: 'Card that triggers an action on the install',
  },
  component: {
    label: 'Deploy component',
    icon: 'CubeIcon',
    hint: 'Card that deploys a component to the install',
  },
  raw: {
    label: 'Template',
    icon: 'BracketsCurlyIcon',
    hint: 'Verbatim Go template escape hatch',
  },
}

const blockGroups: { label: string; types: TBlockType[] }[] = [
  { label: 'Content', types: ['markdown', 'banner', 'status-row', 'table'] },
  { label: 'Interactive', types: ['runbook', 'action', 'component'] },
  { label: 'Advanced', types: ['raw'] },
]

const newBlock = (type: TBlockType): TBlock => {
  const key = crypto.randomUUID()
  switch (type) {
    case 'markdown':
      return { key, type, content: '' }
    case 'banner':
      return { key, type, theme: 'info', content: '' }
    case 'status-row':
      return {
        key,
        type,
        items: [
          { key: crypto.randomUUID(), label: '', kind: 'status', path: '' },
        ],
      }
    case 'table':
      return {
        key,
        type,
        sourcePath: '',
        limit: 5,
        emptyText: 'No data yet',
        columns: [
          { key: crypto.randomUUID(), header: '', kind: 'text', path: '' },
        ],
      }
    case 'runbook':
    case 'action':
    case 'component':
      return { key, type, id: '', name: '' }
    case 'raw':
      return { key, type, content: '' }
  }
}

const draftKey = (appId: string) => `readme-studio-draft-${appId}`

interface IReadmeStudio {
  appId?: string
  embedded?: boolean
  installs?: { id: string; name: string }[]
  runbooks?: TEntityOption[]
  actions?: TEntityOption[]
  components?: TEntityOption[]
  previewInstallId?: string
  previewInstallState?: Record<string, unknown>
  previewInstallStateLoading?: boolean
  onPreviewInstallChange?: (installId: string) => void
  loadingError?: boolean
}

function BlockMenuItems({ onAdd }: { onAdd: (type: TBlockType) => void }) {
  return (
    <Menu className="w-64">
      {blockGroups.flatMap((group) => [
        <Text key={group.label}>{group.label}</Text>,
        ...group.types.map((type) => (
          <Button
            key={type}
            onClick={() => onAdd(type)}
            tooltipProps={{
              position: 'right',
              tipContent: blockMeta[type].hint,
            }}
          >
            <span className="flex items-center gap-2">
              <Icon variant={blockMeta[type].icon} size="14" />
              {blockMeta[type].label}
            </span>
          </Button>
        )),
      ])}
    </Menu>
  )
}

function InsertZone({
  id,
  onAdd,
}: {
  id: string
  onAdd: (type: TBlockType) => void
}) {
  return (
    <div className="group/insert relative flex h-3 -my-1.5 items-center justify-center z-10">
      <div className="absolute inset-x-0 top-1/2 h-px bg-blue-500/40 opacity-0 group-hover/insert:opacity-100 group-focus-within/insert:opacity-100 transition-opacity" />
      <Dropdown
        id={id}
        variant="ghost"
        size="xs"
        hideIcon
        buttonText={<Icon variant="PlusIcon" size="12" />}
        buttonClassName="relative !h-5 !w-5 !p-0 !rounded-full !bg-white dark:!bg-dark-grey-800 border border-cool-grey-300 dark:border-cool-grey-600/40 opacity-0 group-hover/insert:opacity-100 group-focus-within/insert:opacity-100 focus:opacity-100 transition-opacity"
        aria-label="Insert block"
      >
        <BlockMenuItems onAdd={onAdd} />
      </Dropdown>
    </div>
  )
}

function SortableBlock({
  block,
  active,
  preview,
  onActivate,
  onDone,
  onRemove,
  children,
}: {
  block: TBlock
  active: boolean
  preview: string
  onActivate: () => void
  onDone: () => void
  onRemove: () => void
  children: React.ReactNode
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: block.key })

  const isEntityBlock =
    block.type === 'runbook' ||
    block.type === 'action' ||
    block.type === 'component'
  const isEmpty = !preview.trim() || (isEntityBlock && !block.id)

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={cn('group/block relative', isDragging && 'opacity-60 z-20')}
    >
      <div
        className={cn(
          'absolute -left-8 top-1.5 flex flex-col items-center transition-opacity',
          active ? 'opacity-100' : 'opacity-0 group-hover/block:opacity-100'
        )}
      >
        <button
          className="cursor-grab active:cursor-grabbing p-0.5 rounded text-cool-grey-400 hover:text-cool-grey-600 hover:bg-cool-grey-100 dark:hover:text-cool-grey-300 dark:hover:bg-dark-grey-600"
          aria-label={`Drag to reorder ${blockMeta[block.type].label} block`}
          {...attributes}
          {...listeners}
        >
          <Icon variant="DotsSixVerticalIcon" size="14" />
        </button>
      </div>

      {active ? (
        <div className="rounded-lg border border-blue-500/50 ring-1 ring-blue-500/30 bg-white dark:bg-dark-grey-800 shadow-sm">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-cool-grey-200 dark:border-cool-grey-600/30">
            <Icon variant={blockMeta[block.type].icon} size="14" />
            <Text variant="subtext" weight="strong">
              {blockMeta[block.type].label}
            </Text>
            <Text
              variant="subtext"
              theme="neutral"
              className="hidden md:block truncate"
            >
              {blockMeta[block.type].hint}
            </Text>
            <div className="ml-auto flex items-center gap-1">
              <Button
                variant="ghost"
                size="xs"
                aria-label="Remove block"
                onClick={onRemove}
              >
                <Icon variant="TrashIcon" size="14" />
              </Button>
              <Button variant="secondary" size="xs" onClick={onDone}>
                Done
              </Button>
            </div>
          </div>
          <div className="p-3">{children}</div>
        </div>
      ) : (
        <div
          role="button"
          tabIndex={0}
          onClick={onActivate}
          onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              onActivate()
            }
          }}
          className="rounded-md -mx-2 px-2 py-1.5 cursor-pointer transition-colors hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500/50"
        >
          {isEmpty ? (
            <div className="flex items-center gap-2 py-1 text-cool-grey-400 dark:text-cool-grey-500">
              <Icon variant={blockMeta[block.type].icon} size="14" />
              <Text variant="subtext" theme="neutral">
                {isEntityBlock
                  ? `${blockMeta[block.type].label} — click to configure`
                  : `Empty ${blockMeta[block.type].label.toLowerCase()} — click to edit`}
              </Text>
            </div>
          ) : (
            <Markdown content={preview} />
          )}
        </div>
      )}
    </div>
  )
}

export function ReadmeStudio({
  appId = '',
  embedded,
  installs = [],
  runbooks = [],
  actions = [],
  components = [],
  previewInstallId = '',
  previewInstallState,
  previewInstallStateLoading,
  onPreviewInstallChange,
  loadingError,
}: IReadmeStudio) {
  const [blocks, setBlocks] = useState<TBlock[]>(() => {
    if (appId) {
      try {
        const draft = localStorage.getItem(draftKey(appId))
        if (draft) return JSON.parse(draft) as TBlock[]
      } catch {
        /* fall through to seed */
      }
    }
    return seedBlocks()
  })
  const [tab, setTab] = useState<'preview' | 'template'>('preview')
  const [copied, setCopied] = useState(false)
  const [activeKey, setActiveKey] = useState<string | null>(null)

  useEffect(() => {
    if (!appId) return
    const timeout = setTimeout(
      () => localStorage.setItem(draftKey(appId), JSON.stringify(blocks)),
      400
    )
    return () => clearTimeout(timeout)
  }, [appId, blocks])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } })
  )

  const variables = useMemo(
    () => getStateVariables(previewInstallState),
    [previewInstallState]
  )
  const arraySources = useMemo(
    () => getArraySources(previewInstallState),
    [previewInstallState]
  )
  const template = useMemo(() => compileTemplate(blocks), [blocks])
  const preview = useMemo(
    () => previewDocument(blocks, previewInstallState),
    [blocks, previewInstallState]
  )

  const onDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    setBlocks((current) => {
      const from = current.findIndex((block) => block.key === active.id)
      const to = current.findIndex((block) => block.key === over.id)
      return arrayMove(current, from, to)
    })
  }

  const insertAt = (index: number, type: TBlockType) => {
    const block = newBlock(type)
    setBlocks((current) => [
      ...current.slice(0, index),
      block,
      ...current.slice(index),
    ])
    setActiveKey(block.key)
  }

  const removeBlock = (key: string) => {
    setBlocks((current) => current.filter((item) => item.key !== key))
    setActiveKey((current) => (current === key ? null : current))
  }

  const copyTemplate = () => {
    navigator.clipboard.writeText(template)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-3">
        {!embedded ? (
          <div className="flex flex-col">
            <Text weight="strong">README studio</Text>
            <Text variant="subtext" theme="neutral">
              Compose your install README from blocks, preview it with real
              install state, then copy the template into your app config.
            </Text>
          </div>
        ) : null}
        <div className="ml-auto flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={copyTemplate}>
            <Icon variant={copied ? 'CheckIcon' : 'CopyIcon'} size="14" />
            {copied ? 'Copied' : 'Copy template'}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() =>
              downloadFileOnClick({
                content: template,
                filename: 'README.md',
                fileType: 'md',
                mimeType: 'text/markdown',
              })
            }
          >
            <Icon variant="DownloadSimpleIcon" size="14" />
            Download README.md
          </Button>
        </div>
      </div>

      {loadingError ? (
        <Banner theme="warn">
          Install state could not be loaded. You can still build; the preview
          will show placeholders.
        </Banner>
      ) : null}

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 items-start">
        <div className="flex flex-col gap-2">
          <Card className="!py-5 !pl-11 !pr-6 !gap-0">
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={onDragEnd}
            >
              <SortableContext
                items={blocks.map((block) => block.key)}
                strategy={verticalListSortingStrategy}
              >
                {blocks.map((block, index) => (
                  <div key={block.key} className="flex flex-col">
                    <InsertZone
                      id={`insert-${block.key}`}
                      onAdd={(type) => insertAt(index, type)}
                    />
                    <SortableBlock
                      block={block}
                      active={activeKey === block.key}
                      preview={previewBlock(block, previewInstallState)}
                      onActivate={() => setActiveKey(block.key)}
                      onDone={() => setActiveKey(null)}
                      onRemove={() => removeBlock(block.key)}
                    >
                      <BlockEditor
                        block={block}
                        variables={variables}
                        arraySources={arraySources}
                        runbooks={runbooks}
                        actions={actions}
                        components={components}
                        onChange={(next) =>
                          setBlocks((current) =>
                            current.map((item) =>
                              item.key === next.key ? next : item
                            )
                          )
                        }
                      />
                    </SortableBlock>
                  </div>
                ))}
              </SortableContext>
            </DndContext>
            <div className="mt-3 -ml-2">
              <Dropdown
                id="add-block-end"
                variant="ghost"
                size="xs"
                hideIcon
                buttonText={
                  <span className="flex items-center gap-1.5 text-cool-grey-500 dark:text-cool-grey-400">
                    <Icon variant="PlusIcon" size="14" />
                    Add block
                  </span>
                }
              >
                <BlockMenuItems
                  onAdd={(type) => insertAt(blocks.length, type)}
                />
              </Dropdown>
            </div>
          </Card>
          {blocks.length > 0 ? (
            <div className="flex justify-center">
              <Button
                variant="ghost"
                size="xs"
                onClick={() => {
                  setBlocks(seedBlocks())
                  setActiveKey(null)
                }}
              >
                Reset to example
              </Button>
            </div>
          ) : null}
        </div>

        <div className="flex flex-col gap-3 xl:sticky xl:top-4">
          <Card className="!p-3 !gap-3">
            <div className="flex items-center gap-2">
              <div className="inline-flex items-center rounded-full bg-cool-grey-100 dark:bg-dark-grey-700 p-0.5">
                {(['preview', 'template'] as const).map((value) => (
                  <button
                    key={value}
                    className={cn(
                      'rounded-full px-3 py-1 text-xs capitalize transition-colors',
                      tab === value
                        ? 'bg-white dark:bg-dark-grey-500 font-medium shadow-sm'
                        : 'text-cool-grey-500 dark:text-cool-grey-400 hover:text-cool-grey-700 dark:hover:text-cool-grey-200'
                    )}
                    onClick={() => setTab(value)}
                  >
                    {value}
                  </button>
                ))}
              </div>
              <div className="ml-auto w-56">
                <Select
                  size="sm"
                  options={installs.map((install) => ({
                    value: install.id,
                    label: install.name,
                  }))}
                  value={previewInstallId}
                  placeholder="Preview with install data"
                  onChange={(value) => onPreviewInstallChange?.(value)}
                />
              </div>
            </div>
            {previewInstallStateLoading ? (
              <Text variant="subtext" theme="neutral">
                Loading install state...
              </Text>
            ) : null}
            {tab === 'preview' ? (
              <div className="rounded-lg border border-cool-grey-300 dark:border-cool-grey-600/40 p-4 overflow-auto max-h-[70vh]">
                {preview.trim() ? (
                  <Markdown content={preview} />
                ) : (
                  <Text variant="subtext" theme="neutral">
                    No blocks yet. Add a block to see the preview.
                  </Text>
                )}
              </div>
            ) : (
              <div className="overflow-auto max-h-[70vh]">
                <CodeBlock language="markdown">{template}</CodeBlock>
              </div>
            )}
          </Card>
          <div className="flex items-start gap-1.5 px-1">
            <Icon
              variant="InfoIcon"
              size="14"
              className="mt-0.5 shrink-0 text-cool-grey-400"
            />
            <Text variant="subtext" theme="neutral">
              Copy the template into your app config as the install README — it
              renders with live install state on every install page.
            </Text>
          </div>
        </div>
      </div>
    </div>
  )
}
