import { useEffect, useMemo, useRef, useState } from 'react'
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
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Markdown } from '@/components/common/Markdown'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import {
  getStateVariables,
  substituteVariables,
} from '@/components/readme/ReadmeStudio/compiler'
import {
  operationLabels,
  slugifyRunbookName,
  type TBuilderOperation,
  type TBuilderStep,
  type TOption,
} from '@/components/runbooks/RunbookBuilder/helpers'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import { cn } from '@/utils/classnames'
import { downloadFileOnClick } from '@/utils/file-download'
import { StepEditor } from './StepEditor'
import {
  importRunbookCells,
  newMarkdownCell,
  newStepCell,
  notebookToml,
  seedCells,
  serializeNotebookReadme,
  stepTarget,
  validateNotebook,
} from './helpers'
import type { TCell } from './types'

const operationIcons: Record<TBuilderOperation, TIconVariant> = {
  'deploy-component': 'CubeIcon',
  'check-component-drift': 'MagnifyingGlassIcon',
  'tear-down-component': 'TrashIcon',
  'reprovision-sandbox': 'ArrowsClockwiseIcon',
  'check-sandbox-drift': 'MagnifyingGlassIcon',
  'deprovision-sandbox': 'StackIcon',
  'configured-action': 'LightningIcon',
  command: 'TerminalIcon',
}

const operations = Object.keys(operationLabels) as TBuilderOperation[]

const draftKey = (appId: string) => `operations-studio-runbook-${appId}`

type TDraft = { name: string; description: string; cells: TCell[] }

function loadDraft(appId?: string): TDraft {
  if (appId) {
    try {
      const draft = localStorage.getItem(draftKey(appId))
      if (draft) return JSON.parse(draft) as TDraft
    } catch {
      /* fall through to seed */
    }
  }
  return { name: '', description: '', cells: seedCells() }
}

function CellMenuItems({
  onAddMarkdown,
  onAddStep,
}: {
  onAddMarkdown: () => void
  onAddStep: (operation: TBuilderOperation) => void
}) {
  return (
    <Menu className="w-64">
      <Text>Content</Text>
      <Button onClick={onAddMarkdown}>
        <span className="flex items-center gap-2">
          <Icon variant="TextAaIcon" size="14" />
          Markdown
        </span>
      </Button>
      <Text>Executable steps</Text>
      {operations.map((operation) => (
        <Button key={operation} onClick={() => onAddStep(operation)}>
          <span className="flex items-center gap-2">
            <Icon variant={operationIcons[operation]} size="14" />
            {operationLabels[operation]}
          </span>
        </Button>
      ))}
    </Menu>
  )
}

function InsertZone({
  id,
  onAddMarkdown,
  onAddStep,
}: {
  id: string
  onAddMarkdown: () => void
  onAddStep: (operation: TBuilderOperation) => void
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
        aria-label="Insert cell"
      >
        <CellMenuItems onAddMarkdown={onAddMarkdown} onAddStep={onAddStep} />
      </Dropdown>
    </div>
  )
}

function StepSummary({
  step,
  stepNumber,
  issues,
}: {
  step: TBuilderStep
  stepNumber: number
  issues: string[]
}) {
  const target = stepTarget(step)
  return (
    <div className="flex items-center gap-3 min-w-0">
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-500/10 text-blue-700 dark:text-blue-400 text-xs font-semibold">
        {stepNumber}
      </span>
      <Icon
        variant={operationIcons[step.operation]}
        size="16"
        className="shrink-0 text-cool-grey-500 dark:text-cool-grey-400"
      />
      <div className="flex min-w-0 flex-col">
        <Text weight="strong" className="truncate">
          {step.name || operationLabels[step.operation]}
        </Text>
        {step.name && step.name !== operationLabels[step.operation] ? (
          <Text variant="subtext" theme="neutral" className="truncate">
            {operationLabels[step.operation]}
          </Text>
        ) : null}
      </div>
      <div className="ml-auto flex shrink-0 items-center gap-2">
        {target ? (
          <Badge variant="code" size="sm" theme="default" className="max-w-40 truncate">
            {target}
          </Badge>
        ) : null}
        {issues.length ? (
          <Badge size="sm" theme="warn">
            Needs setup
          </Badge>
        ) : null}
      </div>
    </div>
  )
}

function SortableCell({
  cell,
  stepNumber,
  active,
  issues,
  onActivate,
  onDone,
  onRemove,
  children,
}: {
  cell: TCell
  stepNumber?: number
  active: boolean
  issues: string[]
  onActivate: () => void
  onDone: () => void
  onRemove: () => void
  children: React.ReactNode
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: cell.key })
  const isStep = cell.kind === 'step'

  return (
    <div
      ref={setNodeRef}
      id={`cell-${cell.key}`}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={cn('group/cell relative scroll-mt-24', isDragging && 'opacity-60 z-20')}
    >
      <div
        className={cn(
          'absolute -left-8 top-2 flex flex-col items-center transition-opacity',
          active ? 'opacity-100' : 'opacity-0 group-hover/cell:opacity-100'
        )}
      >
        <button
          className="cursor-grab active:cursor-grabbing p-0.5 rounded text-cool-grey-400 hover:text-cool-grey-600 hover:bg-cool-grey-100 dark:hover:text-cool-grey-300 dark:hover:bg-dark-grey-600"
          aria-label={
            isStep ? 'Drag to reorder step' : 'Drag to reorder markdown cell'
          }
          {...attributes}
          {...listeners}
        >
          <Icon variant="DotsSixVerticalIcon" size="14" />
        </button>
      </div>

      {active ? (
        <div className="rounded-lg border border-blue-500/50 ring-1 ring-blue-500/30 bg-white dark:bg-dark-grey-800 shadow-sm">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-cool-grey-200 dark:border-cool-grey-600/30">
            <Icon
              variant={
                isStep ? operationIcons[cell.step.operation] : 'TextAaIcon'
              }
              size="14"
            />
            <Text variant="subtext" weight="strong">
              {isStep
                ? `Step ${stepNumber} · ${operationLabels[cell.step.operation]}`
                : 'Markdown'}
            </Text>
            <div className="ml-auto flex items-center gap-1">
              <Button
                variant="ghost"
                size="xs"
                aria-label={isStep ? 'Remove step' : 'Remove cell'}
                onClick={onRemove}
              >
                <Icon variant="TrashIcon" size="14" />
              </Button>
              <Button variant="secondary" size="xs" onClick={onDone}>
                Done
              </Button>
            </div>
          </div>
          <div className="p-3">
            {children}
            {issues.length ? (
              <div className="mt-3 flex flex-col gap-0.5">
                {issues.map((issue) => (
                  <Text key={issue} variant="subtext" theme="warn">
                    {issue}
                  </Text>
                ))}
              </div>
            ) : null}
          </div>
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
          className={cn(
            'cursor-pointer transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500/50',
            isStep
              ? 'rounded-lg border border-cool-grey-300 dark:border-cool-grey-600/40 bg-cool-grey-50/70 dark:bg-dark-grey-800 px-3 py-2.5 my-1 hover:border-blue-500/50'
              : 'rounded-md -mx-2 px-2 py-1.5 hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700/60'
          )}
        >
          {isStep ? (
            <StepSummary
              step={cell.step}
              stepNumber={stepNumber ?? 0}
              issues={issues}
            />
          ) : cell.content.trim() ? (
            <Markdown content={cell.content} />
          ) : (
            <div className="flex items-center gap-2 py-1 text-cool-grey-400 dark:text-cool-grey-500">
              <Icon variant="TextAaIcon" size="14" />
              <Text variant="subtext" theme="neutral">
                Empty markdown — click to edit
              </Text>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function PreviewStepCard({
  step,
  stepNumber,
}: {
  step: TBuilderStep
  stepNumber: number
}) {
  const target = stepTarget(step)
  return (
    <div className="flex items-center gap-3 rounded-lg border border-cool-grey-300 dark:border-cool-grey-600/40 bg-cool-grey-50/70 dark:bg-dark-grey-800 px-4 py-3">
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-500/10 text-blue-700 dark:text-blue-400 text-xs font-semibold">
        {stepNumber}
      </span>
      <Icon
        variant={operationIcons[step.operation]}
        size="16"
        className="shrink-0 text-cool-grey-500 dark:text-cool-grey-400"
      />
      <div className="flex min-w-0 flex-col">
        <Text weight="strong" className="truncate">
          {step.name || operationLabels[step.operation]}
        </Text>
        <Text variant="subtext" theme="neutral" className="truncate">
          {operationLabels[step.operation]}
        </Text>
      </div>
      <div className="ml-auto flex shrink-0 items-center gap-2">
        {target ? (
          <Badge variant="code" size="sm" theme="default" className="max-w-40 truncate">
            {target}
          </Badge>
        ) : null}
        <Button variant="secondary" size="xs" disabled>
          <Icon variant="PlayIcon" size="12" />
          Run
        </Button>
      </div>
    </div>
  )
}

function VariablesPanel({
  state,
  onInsert,
}: {
  state?: Record<string, unknown>
  onInsert: (template: string) => 'inserted' | 'copied'
}) {
  const [query, setQuery] = useState('')
  const [feedback, setFeedback] = useState<{
    template: string
    result: 'inserted' | 'copied'
  }>()
  const feedbackTimer = useRef<number | undefined>(undefined)
  useEffect(() => () => window.clearTimeout(feedbackTimer.current), [])
  const variables = useMemo(() => getStateVariables(state), [state])
  const filtered = variables.filter(
    (variable) =>
      variable.template.toLowerCase().includes(query.toLowerCase()) ||
      (variable.value ?? '').toLowerCase().includes(query.toLowerCase())
  )

  return (
    <Card className="!p-4 !gap-3">
      <div className="flex items-center gap-2">
        <Icon variant="BracketsCurlyIcon" size="14" />
        <Text weight="strong">Install state variables</Text>
        <Text variant="subtext" theme="neutral" className="ml-auto">
          {filtered.length} of {variables.length}
        </Text>
      </div>
      <Input
        placeholder="Search variables"
        aria-label="Search state variables"
        value={query}
        onChange={(event) => setQuery(event.target.value)}
      />
      <div className="flex max-h-72 flex-col gap-0.5 overflow-y-auto">
        {!filtered.length ? (
          <Text variant="subtext" theme="neutral">
            No variables match "{query}".
          </Text>
        ) : (
          filtered.map((variable) => (
            <button
              key={variable.template}
              className="flex items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700/60"
              onClick={() => {
                const result = onInsert(variable.template)
                setFeedback({ template: variable.template, result })
                window.clearTimeout(feedbackTimer.current)
                feedbackTimer.current = window.setTimeout(() => setFeedback(undefined), 1500)
              }}
            >
              <code className="shrink-0 text-xs text-blue-700 dark:text-blue-400">
                {variable.template}
              </code>
              <Text
                variant="subtext"
                theme="neutral"
                className="ml-auto truncate max-w-40"
              >
                {feedback?.template === variable.template
                  ? feedback.result === 'inserted'
                    ? 'Inserted'
                    : 'Copied'
                  : (variable.value ?? '—')}
              </Text>
            </button>
          ))
        )}
      </div>
      <Text variant="subtext" theme="neutral">
        Click to insert into the active markdown cell, or copy when no cell is
        being edited. Values shown from the preview install.
      </Text>
    </Card>
  )
}

interface IRunbookNotebook {
  appId?: string
  components: TOption[]
  actions: TOption[]
  runbooks: TRunbook[]
  installs?: TOption[]
  previewInstallId?: string
  previewInstallState?: Record<string, unknown>
  previewInstallStateLoading?: boolean
  onPreviewInstallChange?: (installId: string) => void
  loading?: boolean
  loadingError?: boolean
}

export function RunbookNotebook({
  appId = '',
  components,
  actions,
  runbooks,
  installs = [],
  previewInstallId = '',
  previewInstallState,
  previewInstallStateLoading,
  onPreviewInstallChange,
  loading,
  loadingError,
}: IRunbookNotebook) {
  const [draft] = useState(() => loadDraft(appId))
  const [name, setName] = useState(draft.name)
  const [description, setDescription] = useState(draft.description)
  const [cells, setCells] = useState<TCell[]>(draft.cells)
  const [activeKey, setActiveKey] = useState<string | null>(null)
  const [importId, setImportId] = useState('')
  const [messages, setMessages] = useState<string[]>([])
  const [copied, setCopied] = useState<'toml' | 'markdown'>()
  const copiedTimer = useRef<number | undefined>(undefined)
  useEffect(() => () => window.clearTimeout(copiedTimer.current), [])
  const [previewTab, setPreviewTab] = useState<'preview' | 'toml' | 'template'>(
    'preview'
  )

  useEffect(() => {
    if (!appId) return
    const timeout = setTimeout(
      () =>
        localStorage.setItem(
          draftKey(appId),
          JSON.stringify({ name, description, cells })
        ),
      400
    )
    return () => clearTimeout(timeout)
  }, [appId, name, description, cells])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } })
  )

  const validation = useMemo(
    () => validateNotebook(name, cells),
    [name, cells]
  )
  const toml = useMemo(
    () => notebookToml(name, description, cells),
    [name, description, cells]
  )
  const markdown = useMemo(
    () => serializeNotebookReadme(name, description, cells),
    [name, description, cells]
  )
  const slug = slugifyRunbookName(name)

  const stepNumbers = useMemo(() => {
    const numbers = new Map<string, number>()
    let index = 0
    cells.forEach((cell) => {
      if (cell.kind === 'step') numbers.set(cell.key, ++index)
    })
    return numbers
  }, [cells])
  const stepCells = cells.filter((cell) => cell.kind === 'step')

  const onDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    setCells((current) => {
      const from = current.findIndex((cell) => cell.key === active.id)
      const to = current.findIndex((cell) => cell.key === over.id)
      return arrayMove(current, from, to)
    })
  }

  const insertAt = (index: number, cell: TCell) => {
    setCells((current) => [
      ...current.slice(0, index),
      cell,
      ...current.slice(index),
    ])
    setActiveKey(cell.key)
  }

  const removeCell = (key: string) => {
    setCells((current) => current.filter((cell) => cell.key !== key))
    setActiveKey((current) => (current === key ? null : current))
  }

  const updateStep = (key: string, patch: Partial<TBuilderStep>) =>
    setCells((current) =>
      current.map((cell) =>
        cell.key === key && cell.kind === 'step'
          ? { ...cell, step: { ...cell.step, ...patch } }
          : cell
      )
    )

  const importSelected = () => {
    const runbook = runbooks.find((item) => item.id === importId)
    if (!runbook) return
    const result = importRunbookCells(runbook, actions)
    setCells((current) => [...current, ...result.cells])
    setMessages(
      result.errors.length
        ? result.errors
        : [`Imported ${result.cells.filter((cell) => cell.kind === 'step').length} steps from ${runbook.name}.`]
    )
  }

  const copy = async (kind: 'toml' | 'markdown') => {
    await navigator.clipboard.writeText(kind === 'toml' ? toml : markdown)
    setCopied(kind)
    window.clearTimeout(copiedTimer.current)
    copiedTimer.current = window.setTimeout(() => setCopied(undefined), 1500)
  }

  const insertVariable = (template: string): 'inserted' | 'copied' => {
    const active = cells.find(
      (cell) => cell.key === activeKey && cell.kind === 'markdown'
    )
    if (!active) {
      void navigator.clipboard.writeText(template)
      return 'copied'
    }
    setCells((current) =>
      current.map((cell) =>
        cell.key === active.key && cell.kind === 'markdown'
          ? {
              ...cell,
              content: cell.content
                ? `${cell.content.replace(/\s+$/, '')} ${template}`
                : template,
            }
          : cell
      )
    )
    return 'inserted'
  }

  return (
    <div className="flex flex-col gap-4">
      {loadingError ? (
        <Banner theme="warn">
          Some app resources could not be loaded. You can still build with
          available options.
        </Banner>
      ) : null}
      {messages.map((message) => (
        <Banner
          key={message}
          theme={message.startsWith('Imported') ? 'success' : 'warn'}
        >
          {message}
        </Banner>
      ))}

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 items-start">
        <div className="flex flex-col gap-3">
          <Card className="!py-6 !pl-11 !pr-6 !gap-0">
            <input
              className="w-full bg-transparent text-2xl font-semibold outline-none placeholder:text-cool-grey-400 dark:placeholder:text-cool-grey-600"
              placeholder="Untitled runbook"
              aria-label="Runbook name"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
            <input
              className="w-full bg-transparent text-sm outline-none mt-1 mb-4 text-cool-grey-700 dark:text-cool-grey-300 placeholder:text-cool-grey-400 dark:placeholder:text-cool-grey-600"
              placeholder="Add a description"
              aria-label="Runbook description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={onDragEnd}
            >
              <SortableContext
                items={cells.map((cell) => cell.key)}
                strategy={verticalListSortingStrategy}
              >
                {cells.map((cell, index) => (
                  <div key={cell.key} className="flex flex-col">
                    <InsertZone
                      id={`insert-${cell.key}`}
                      onAddMarkdown={() => insertAt(index, newMarkdownCell())}
                      onAddStep={(operation) =>
                        insertAt(index, newStepCell(operation))
                      }
                    />
                    <SortableCell
                      cell={cell}
                      stepNumber={stepNumbers.get(cell.key)}
                      active={activeKey === cell.key}
                      issues={validation.byCell[cell.key] ?? []}
                      onActivate={() => setActiveKey(cell.key)}
                      onDone={() => setActiveKey(null)}
                      onRemove={() => removeCell(cell.key)}
                    >
                      {cell.kind === 'markdown' ? (
                        <Textarea
                          id={`markdown-${cell.key}`}
                          aria-label="Markdown content"
                          placeholder="Explain what happens next, add checklists, warnings, or links."
                          value={cell.content}
                          autoResize
                          minRows={3}
                          onChange={(event) =>
                            setCells((current) =>
                              current.map((item) =>
                                item.key === cell.key &&
                                item.kind === 'markdown'
                                  ? { ...item, content: event.target.value }
                                  : item
                              )
                            )
                          }
                        />
                      ) : (
                        <StepEditor
                          step={cell.step}
                          components={components}
                          actions={actions}
                          onChange={(patch) => updateStep(cell.key, patch)}
                        />
                      )}
                    </SortableCell>
                  </div>
                ))}
              </SortableContext>
            </DndContext>
            <div className="mt-3 -ml-2">
              <Dropdown
                id="add-cell-end"
                variant="ghost"
                size="xs"
                hideIcon
                buttonText={
                  <span className="flex items-center gap-1.5 text-cool-grey-500 dark:text-cool-grey-400">
                    <Icon variant="PlusIcon" size="14" />
                    Add content or step
                  </span>
                }
              >
                <CellMenuItems
                  onAddMarkdown={() =>
                    insertAt(cells.length, newMarkdownCell())
                  }
                  onAddStep={(operation) =>
                    insertAt(cells.length, newStepCell(operation))
                  }
                />
              </Dropdown>
            </div>
          </Card>

          <VariablesPanel
            state={previewInstallState}
            onInsert={insertVariable}
          />

          <div className="flex items-center gap-2">
            <Select
              size="sm"
              className="max-w-64"
              options={runbooks.map((runbook) => ({
                value: runbook.id,
                label: runbook.name,
              }))}
              value={importId}
              placeholder={
                loading ? 'Loading runbooks' : 'Start from an existing runbook'
              }
              onChange={(event) => setImportId(event.target.value)}
              disabled={loading || !runbooks.length}
            />
            <Button
              variant="secondary"
              size="sm"
              disabled={!importId}
              onClick={importSelected}
            >
              Import steps
            </Button>
            <Button
              variant="ghost"
              size="xs"
              className="ml-auto"
              onClick={() => {
                setName('')
                setDescription('')
                setCells(seedCells())
                setActiveKey(null)
              }}
            >
              Reset to example
            </Button>
          </div>
        </div>

        <div className="flex flex-col gap-3 xl:sticky xl:top-4">
          <Card className="!p-0 !gap-0 overflow-hidden">
            <div className="flex flex-wrap items-center gap-3 border-b border-cool-grey-200 dark:border-cool-grey-600/30 px-4 py-2.5">
              <div className="inline-flex items-center rounded-full bg-cool-grey-100 dark:bg-dark-grey-700 p-0.5">
                {(
                  [
                    { value: 'preview', label: 'Preview', icon: 'EyeIcon' },
                    { value: 'toml', label: `${slug}.toml`, icon: 'FileCodeIcon' },
                    { value: 'template', label: `${slug}.md`, icon: 'FileTextIcon' },
                  ] as const
                ).map((tab) => (
                  <button
                    key={tab.value}
                    className={cn(
                      'flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs transition-colors',
                      previewTab === tab.value
                        ? 'bg-white dark:bg-dark-grey-500 font-medium shadow-sm'
                        : 'text-cool-grey-500 dark:text-cool-grey-400 hover:text-cool-grey-700 dark:hover:text-cool-grey-200'
                    )}
                    onClick={() => setPreviewTab(tab.value)}
                  >
                    <Icon variant={tab.icon} size="14" />
                    <span className="max-w-36 truncate">{tab.label}</span>
                  </button>
                ))}
              </div>
              <div className="ml-auto flex items-center gap-2">
                {onPreviewInstallChange ? (
                  <Select
                    size="sm"
                    className="max-w-44"
                    aria-label="Preview install"
                    options={installs.map((install) => ({
                      value: install.id,
                      label: install.name,
                    }))}
                    value={previewInstallId}
                    placeholder={
                      previewInstallStateLoading
                        ? 'Loading state'
                        : 'Preview install'
                    }
                    onChange={(event) =>
                      onPreviewInstallChange(event.target.value)
                    }
                    disabled={!installs.length}
                  />
                ) : null}
                {validation.errors.length ? (
                  <>
                    <Icon
                      variant="WarningIcon"
                      size="14"
                      className="text-orange-500"
                    />
                    <Text variant="subtext" theme="warn">
                      {validation.errors.length}{' '}
                      {validation.errors.length === 1 ? 'issue' : 'issues'}
                    </Text>
                  </>
                ) : (
                  <>
                    <Icon
                      variant="CheckIcon"
                      size="14"
                      className="text-green-600"
                    />
                    <Text variant="subtext" theme="success">
                      Ready
                    </Text>
                  </>
                )}
              </div>
            </div>

            <div className="max-h-[calc(100vh-9rem)] overflow-y-auto p-5">
              {previewTab === 'preview' ? (
                <div className="flex flex-col gap-4">
                  <div className="flex flex-col gap-1">
                    <span className="text-2xl font-semibold">
                      {name.trim() || 'Untitled runbook'}
                    </span>
                    {description.trim() ? (
                      <Text theme="neutral">{description.trim()}</Text>
                    ) : null}
                  </div>
                  {cells.map((cell) =>
                    cell.kind === 'markdown' ? (
                      cell.content.trim() ? (
                        <Markdown
                          key={cell.key}
                          content={substituteVariables(
                            cell.content.trim(),
                            previewInstallState
                          )}
                        />
                      ) : null
                    ) : (
                      <PreviewStepCard
                        key={cell.key}
                        step={cell.step}
                        stepNumber={stepNumbers.get(cell.key) ?? 0}
                      />
                    )
                  )}
                  {!cells.length ? (
                    <Text variant="subtext" theme="neutral">
                      Add markdown and steps on the left to see the
                      operator-facing document here.
                    </Text>
                  ) : null}
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <div className="flex items-center gap-2">
                    <Icon
                      variant={
                        previewTab === 'toml' ? 'FileCodeIcon' : 'FileTextIcon'
                      }
                      size="14"
                      className="shrink-0"
                    />
                    <Text variant="subtext" className="truncate">
                      runbooks/{slug}.{previewTab === 'toml' ? 'toml' : 'md'}
                    </Text>
                    <div className="ml-auto flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="xs"
                        aria-label={
                          previewTab === 'toml'
                            ? 'Copy runbook TOML'
                            : 'Copy runbook README'
                        }
                        disabled={!!validation.errors.length}
                        onClick={() =>
                          copy(previewTab === 'toml' ? 'toml' : 'markdown')
                        }
                      >
                        <Icon
                          variant={
                            copied ===
                            (previewTab === 'toml' ? 'toml' : 'markdown')
                              ? 'CheckIcon'
                              : 'CopyIcon'
                          }
                          size="14"
                        />
                      </Button>
                      <Button
                        variant="ghost"
                        size="xs"
                        aria-label={
                          previewTab === 'toml'
                            ? 'Download runbook TOML'
                            : 'Download runbook README'
                        }
                        disabled={!!validation.errors.length}
                        onClick={() =>
                          downloadFileOnClick({
                            content: previewTab === 'toml' ? toml : markdown,
                            filename: `${slug}.${previewTab === 'toml' ? 'toml' : 'md'}`,
                            ...(previewTab === 'template'
                              ? { mimeType: 'text/markdown' }
                              : {}),
                          })
                        }
                      >
                        <Icon variant="DownloadSimpleIcon" size="14" />
                      </Button>
                    </div>
                  </div>
                  <CodeBlock
                    language={previewTab === 'toml' ? 'toml' : 'markdown'}
                    showCopy={false}
                  >
                    {previewTab === 'toml' ? toml : markdown}
                  </CodeBlock>
                  <Text variant="subtext" theme="neutral">
                    Copy this file into your app config's runbooks directory.
                    {previewTab === 'template'
                      ? ' Variables stay as {{.nuon.*}} templates and resolve per install at runtime.'
                      : ''}
                  </Text>
                </div>
              )}
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
