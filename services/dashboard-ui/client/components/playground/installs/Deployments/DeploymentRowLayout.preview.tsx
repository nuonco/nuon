import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { DateTime } from 'luxon'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { DeploymentRow } from '@/components/installs/DeploymentsList/DeploymentRow'
import {
  DeploymentsListFilters,
  type IDeploymentFilter,
} from '@/components/installs/DeploymentsList/DeploymentsListPresenter'
import { ListPage } from '@/components/layout/ListPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useTheme } from '@/hooks/use-theme'
import type { TInstallDeploymentRecordType, TWorkflowStep } from '@/types'
import {
  datePresetQueryParameter,
  WORKFLOW_STATUS_GROUPS,
  type TWorkflowDatePreset,
} from '@/utils/workflow-filters'
import { ResourceOutcomes } from '@/components/installs/DeploymentDetail/DeploymentProgress'
import type { TDeploymentOutcome } from '@/components/installs/DeploymentDetail/deployment-progress'

const scenarios = {
  success: {
    label: 'Success',
    title: 'Roll out template v12',
    type: 'app_branch_update',
    hoursAgo: 48,
    status: 'success',
    activity: 'All updates completed',
  },
  error: {
    label: 'Failure before apply',
    title: 'Update stack permissions',
    type: 'stack_update',
    hoursAgo: 26,
    status: 'error',
    activity: 'Workflow stopped: connection refused',
  },
  partial: {
    label: 'Partial rollout',
    title: 'Roll out template v13',
    type: 'app_branch_update',
    hoursAgo: 10,
    status: 'error',
    activity: 'worker: post-deploy readiness check failed',
  },
  running: {
    label: 'Running',
    title: 'Deploy production components',
    type: 'component_deploy',
    hoursAgo: 1,
    status: 'in-progress',
    activity: 'Deploy components',
  },
  approval: {
    label: 'Pending approval',
    title: 'Roll out template v14',
    type: 'app_branch_update',
    hoursAgo: 2,
    status: 'in-progress',
    activity: 'Waiting for component plan approval',
  },
} satisfies Record<
  string,
  {
    label: string
    title: string
    type: TInstallDeploymentRecordType
    hoursAgo: number
    status: string
    activity: string
  }
>

const emptyFilter: IDeploymentFilter = {
  search: '',
  status: new Set(),
  type: new Set(),
}
type TScenario = keyof typeof scenarios
type TScale = 'auto' | 'comfortable' | 'compact'
type TResourceShare = '35' | '40' | '45'
type TLayoutRun = {
  id: string
  title: string
  type: TInstallDeploymentRecordType
  created_at: string
  status: string
  activity: string
  outcomes: TDeploymentOutcome[]
  steps: TWorkflowStep[]
}
interface IPreviewStyle extends CSSProperties {
  '--workflow-share': string
  '--resource-share': string
}

function makeRun(
  scenario: TScenario,
  title: string,
  activity: string,
  count: number
): TLayoutRun {
  const categories: TDeploymentOutcome['category'][] = [
    'Stack',
    'Sandbox',
    'Components',
  ]
  const outcomes: TDeploymentOutcome[] = categories
    .slice(0, count)
    .flatMap((category, index) => {
      const status =
        scenario === 'success'
          ? 'success'
          : scenario === 'error'
            ? index === 0
              ? 'error'
              : 'not-started'
            : scenario === 'partial'
              ? index === count - 1
                ? 'error'
                : 'success'
              : scenario === 'running'
                ? index === count - 1
                  ? 'in-progress'
                  : 'success'
                : 'not-started'
      const outcome = {
        category,
        name: category,
        status,
        detail:
          status === 'success'
            ? 'Update completed'
            : status === 'error'
              ? 'Update failed'
              : status === 'in-progress'
                ? 'Update running'
                : 'Not started',
      }
      return scenario === 'partial' && index === count - 1
        ? [
            {
              ...outcome,
              name: category === 'Components' ? 'api' : `${category} previous`,
              status: 'success',
              detail: 'Rollout completed',
            },
            {
              ...outcome,
              name:
                category === 'Components' ? 'worker' : `${category} updated`,
              detail: 'Applied; readiness check failed',
            },
          ]
        : [outcome]
    })
  const steps: TWorkflowStep[] = [
    {
      id: 'layout-sync',
      name: 'Sync configuration',
      status: { status: 'success' },
    },
    {
      id: 'layout-apply',
      name:
        scenario === 'error'
          ? 'Plan stack permissions'
          : scenario === 'running' || scenario === 'approval'
            ? activity
            : 'Deploy components',
      status: {
        status:
          scenario === 'approval'
            ? 'approval-awaiting'
            : scenario === 'running'
              ? 'in-progress'
              : scenario === 'error'
                ? 'error'
                : 'success',
      },
    },
    {
      id: 'layout-health',
      name:
        scenario === 'error'
          ? 'Apply stack policy'
          : scenario === 'partial'
            ? 'Verify component readiness'
            : 'Verify install health',
      status: {
        status:
          scenario === 'success'
            ? 'success'
            : scenario === 'partial'
              ? 'error'
              : 'pending',
      },
    },
  ]
  if (scenario === 'partial')
    steps.push({
      id: 'layout-finalize',
      name: 'Finalize deployment',
      status: { status: 'pending' },
    })
  return {
    id: `layout-${scenario}`,
    title,
    type: scenarios[scenario].type,
    created_at: DateTime.now()
      .minus({ hours: scenarios[scenario].hoursAgo })
      .toISO(),
    activity,
    outcomes,
    status: scenarios[scenario].status,
    steps,
  }
}

const LayoutDetails = ({ run, ...props }: IPanel & { run: TLayoutRun }) => (
  <Panel
    {...props}
    heading={run.title}
    size="3/4"
    aria-label="Layout preview details"
  >
    <Text variant="subtext" theme="neutral">
      Fixture data only · Preview layout, not API behavior
    </Text>
    <Text variant="base" theme={run.status === 'error' ? 'error' : 'neutral'}>
      {run.activity}
    </Text>
    <ResourceOutcomes run={run} />
  </Panel>
)

const LayoutRow = ({ run, scale }: { run: TLayoutRun; scale: TScale }) => {
  const { addPanel } = useSurfaces()
  return (
    <DeploymentRow
      run={run}
      title={run.title}
      createdAt={run.created_at}
      href={`/deployments/${run.id}`}
      scale={scale}
      onViewDetails={() => addPanel(<LayoutDetails run={run} />)}
    />
  )
}

export const LayoutControls = () => {
  const [scenario, setScenario] = useState<TScenario>('running')
  const [title, setTitle] = useState('Provisioned install')
  const [activity, setActivity] = useState(scenarios.running.activity)
  const [count, setCount] = useState('3')
  const [workflowCount, setWorkflowCount] = useState('5')
  const [filter, setFilter] = useState<IDeploymentFilter>(emptyFilter)
  const [width, setWidth] = useState('available')
  const [resourceShare, setResourceShare] = useState<TResourceShare>('40')
  const [scale, setScale] = useState<TScale>('auto')
  const [measuredWidth, setMeasuredWidth] = useState(0)
  const preview = useRef<HTMLElement>(null)
  const { preference, setPreference } = useTheme()
  const previewStyle: IPreviewStyle = {
    width: width === 'available' ? '100%' : Number(width),
    '--workflow-share': `${100 - Number(resourceShare)}fr`,
    '--resource-share': `${resourceShare}fr`,
  }
  useEffect(() => {
    const element = preview.current
    if (!element) return
    const observer = new ResizeObserver(([entry]) =>
      setMeasuredWidth(Math.round(entry.contentRect.width))
    )
    observer.observe(element)
    return () => observer.disconnect()
  }, [])
  const companionScenarios: TScenario[] = [
    'running',
    'approval',
    'partial',
    'error',
    'success',
  ]
  const allRuns = [
    makeRun(scenario, title, activity, Number(count)),
    ...companionScenarios
      .filter((value) => value !== scenario)
      .map((value) =>
        makeRun(
          value,
          scenarios[value].title,
          scenarios[value].activity,
          value === 'error' ? 1 : 3
        )
      ),
  ].slice(0, Number(workflowCount))
  const resources = [
    ...new Set(
      allRuns.flatMap((run) =>
        run.outcomes.map((outcome) => outcome.category.toLowerCase())
      )
    ),
  ]
  const since = datePresetQueryParameter(filter.date ?? null)
  const runs = allRuns.filter(
    (run) =>
      `${run.title} ${run.activity}`
        .toLowerCase()
        .includes(filter.search.toLowerCase()) &&
      (!filter.status.size ||
        [...filter.status].some((status) =>
          WORKFLOW_STATUS_GROUPS[status].includes(run.status)
        )) &&
      (!filter.type.size || filter.type.has(run.type)) &&
      (!filter.resource ||
        run.outcomes.some(
          (outcome) => outcome.category.toLowerCase() === filter.resource
        )) &&
      (!since || DateTime.fromISO(run.created_at) >= DateTime.fromISO(since))
  )
  return (
    <div className="density-story flex min-w-0 flex-col gap-6 p-4 md:p-6">
      <SectionHeader
        title="Deployment row layout"
        description="Preview the same rows used on the Deployments page. Controls edit the first workflow; companion rows show the other situations. Data and actions are fixtures only."
      />
      <section
        aria-label="Layout controls"
        className="grid grid-cols-1 gap-4 border-b pb-4 sm:grid-cols-2 lg:grid-cols-3"
      >
        <Select
          id="layout-scenario"
          labelProps={{ labelText: 'Workflow situation' }}
          value={scenario}
          options={Object.entries(scenarios).map(([value, entry]) => ({
            value,
            label: entry.label,
          }))}
          onChange={(value) => {
            if (
              value === 'success' ||
              value === 'error' ||
              value === 'partial' ||
              value === 'running' ||
              value === 'approval'
            ) {
              setScenario(value)
              setActivity(scenarios[value].activity)
            }
          }}
        />
        <Select
          id="layout-resource-count"
          labelProps={{ labelText: 'Resource categories' }}
          value={count}
          options={['1', '2', '3'].map((value) => ({
            value,
            label: `${value} ${value === '1' ? 'category' : 'categories'}`,
          }))}
          onChange={setCount}
        />
        <Select
          id="layout-workflow-count"
          labelProps={{ labelText: 'Workflow list' }}
          value={workflowCount}
          options={[
            { value: '1', label: 'Single workflow' },
            { value: '3', label: '3 workflows' },
            { value: '5', label: '5 workflows (all situations)' },
          ]}
          onChange={setWorkflowCount}
        />
        <Select
          id="layout-width"
          labelProps={{ labelText: 'Preview width' }}
          value={width}
          options={[
            'available',
            '390',
            '768',
            '1024',
            '1440',
            '1920',
            '2560',
          ].map((value) => ({
            value,
            label:
              value === 'available' ? 'Available canvas width' : `${value}px`,
          }))}
          onChange={setWidth}
          helperText="Wider presets are limited by your browser's available canvas."
        />
        <Select
          id="layout-column-balance"
          labelProps={{ labelText: 'Column balance' }}
          value={resourceShare}
          options={[
            { value: '35', label: 'Resources 35% / Workflow 65%' },
            { value: '40', label: 'Resources 40% / Workflow 60%' },
            { value: '45', label: 'Resources 45% / Workflow 55%' },
          ]}
          onChange={(value) => {
            if (value === '35' || value === '40' || value === '45')
              setResourceShare(value)
          }}
          helperText="Shares of the space after View details. Resources grow with the row width; the balance stays the same for every workflow."
        />
        <Select
          id="layout-scale"
          labelProps={{ labelText: 'Outcome sizing' }}
          value={scale}
          options={[
            { value: 'auto', label: 'Automatic (fit each sector)' },
            { value: 'comfortable', label: 'Comfortable (18px / 22px icons)' },
            { value: 'compact', label: 'Compact (14px / 14px icons)' },
          ]}
          onChange={(value) => {
            if (
              value === 'auto' ||
              value === 'comfortable' ||
              value === 'compact'
            )
              setScale(value)
          }}
        />
        <Select
          id="layout-theme"
          labelProps={{ labelText: 'Theme' }}
          value={preference}
          options={[
            { value: 'system', label: 'System' },
            { value: 'light', label: 'Light' },
            { value: 'dark', label: 'Dark' },
          ]}
          onChange={(value) => {
            if (value === 'system' || value === 'light' || value === 'dark')
              setPreference(value)
          }}
        />
        <Input
          id="layout-title"
          labelProps={{ labelText: 'Deployment title' }}
          value={title}
          onChange={(event) => setTitle(event.target.value)}
        />
        <div className="sm:col-span-2 lg:col-span-3">
          <Textarea
            id="layout-activity"
            labelProps={{ labelText: 'Workflow description / failure reason' }}
            value={activity}
            onChange={(event) => setActivity(event.target.value)}
            autoResize
            minRows={2}
            maxRows={6}
            helperText="All workflows use the same column widths. Current and next steps remain visible in every state; long failure reasons wrap without moving the resource column."
          />
        </div>
        <Button
          variant="ghost"
          onClick={() => {
            setScenario('running')
            setTitle('Provisioned install')
            setActivity(scenarios.running.activity)
            setCount('3')
            setWorkflowCount('5')
            setFilter(emptyFilter)
            setWidth('available')
            setResourceShare('40')
            setScale('auto')
          }}
        >
          Reset preview
        </Button>
      </section>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Text variant="base" weight="strong">
          {workflowCount === '1'
            ? scenarios[scenario].label
            : 'Mixed workflows'}
        </Text>
        <div role="status">
          <Text variant="subtext" theme="neutral">
            Actual preview width: {measuredWidth}px · {runs.length} of{' '}
            {allRuns.length} workflows ·{' '}
            {scale === 'auto'
              ? 'Automatic sizing: 14–18px text, 14–22px icons'
              : `${scale} sizing`}
          </Text>
        </div>
      </div>
      <main
        ref={preview}
        aria-label="Deployments layout preview"
        className="min-w-0 max-w-full"
        style={previewStyle}
      >
        <ListPage
          title="Deployments"
          description="Follow rollouts for acme-preview. View details to inspect a workflow or its changes."
        >
          <DeploymentsListFilters
            resources={resources}
            search={filter.search}
            filter={filter}
            onSearchChange={(search) =>
              setFilter((current) => ({ ...current, search }))
            }
            onStatusChange={(status) =>
              setFilter((current) => ({ ...current, status }))
            }
            onTypeChange={(type) =>
              setFilter((current) => ({ ...current, type }))
            }
            onResourceChange={(resource) =>
              setFilter((current) => ({ ...current, resource }))
            }
            onDateChange={(date) =>
              setFilter((current) => ({
                ...current,
                date: date as TWorkflowDatePreset | undefined,
              }))
            }
            onClearFilters={() => setFilter(emptyFilter)}
          />
          {runs.length ? (
            <Timeline
              className="w-full"
              events={runs}
              groupByDate={false}
              pagination={{ hasNext: false, offset: 0, limit: 20 }}
              getEventKey={(run) => run.id}
              renderEvent={(run) => <LayoutRow run={run} scale={scale} />}
            />
          ) : (
            <EmptyState
              emptyTitle="No deployments found"
              emptyMessage="No deployments match the current filters. Try adjusting or clearing them."
              className="my-12"
            />
          )}
        </ListPage>
      </main>
    </div>
  )
}
