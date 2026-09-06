import { DateTime } from 'luxon'
import { useState } from 'react'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { Badge } from '../../atoms/Badge'
import { Button } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Icon } from '../../atoms/Icon'
import { Link } from '../../atoms/Link'
import { Text } from '../../atoms/Text'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { TimelineItem } from '../../molecules/TimelineItem'
import { Panel } from '../surfaces/Panel'
import { Timeline } from './Timeline'

export default {
  title: 'lite/organisms/Timeline',
}

type TWorkflowEvent = {
  id: string
  name: string
  status: string
  createdAt: string
  createdBy: string
  commit?: string
  note?: string
  drift?: boolean
}

const minutesAgo = (minutes: number) =>
  DateTime.now().minus({ minutes }).toISO()

const WORKFLOWS: TWorkflowEvent[] = [
  {
    id: 'wfq7fplr1up5atx5zpxotbabm',
    name: 'Payments API deploy',
    status: 'in-progress',
    createdAt: minutesAgo(3),
    createdBy: 'engineer@example.com',
    commit: '4f9c1ab',
  },
  {
    id: 'wfk2mn8xqz4td9wlv6cshjy10',
    name: 'Payments API build',
    status: 'active',
    createdAt: minutesAgo(46),
    createdBy: 'engineer@example.com',
    commit: '4f9c1ab',
  },
  {
    id: 'wfp5rt3bnk8ce7yqda2gxsw94',
    name: 'Dashboard deploy',
    status: 'failed',
    createdAt: minutesAgo(60 * 9),
    createdBy: 'engineer@example.com',
    note: 'Terraform apply exited with status 1.',
  },
  {
    id: 'wfz8dl6hvc1qm4xteo9nrbky7',
    name: 'Sandbox reprovision',
    status: 'active',
    createdAt: minutesAgo(60 * 27),
    createdBy: 'a service account',
    drift: true,
  },
  {
    id: 'wfj4gw9syb2fn6prlk8duqhx3',
    name: 'Dashboard teardown',
    status: 'cancelled',
    createdAt: minutesAgo(60 * 30),
    createdBy: 'engineer@example.com',
  },
  {
    id: 'wfv6cx2qmt8kd5wnyhr7bealj',
    name: 'Payments API deploy',
    status: 'active',
    createdAt: minutesAgo(60 * 74),
    createdBy: 'engineer@example.com',
    commit: '81be0d2',
  },
]

const WorkflowActions = ({ event }: { event: TWorkflowEvent }) =>
  event.status === 'in-progress' ? (
    <Button size="sm" variant="ghost">
      Cancel
    </Button>
  ) : (
    <Button
      size="sm"
      variant="ghost"
      iconOnly
      aria-label={`${event.name} actions`}
      tooltip="Workflow actions"
    >
      <Icon variant="DotsThreeIcon" size={16} aria-hidden />
    </Button>
  )

const WorkflowsTimeline = ({
  events = WORKFLOWS,
  loading,
  group,
  emptyState,
}: {
  events?: TWorkflowEvent[]
  loading?: boolean
  group?: boolean
  emptyState?: React.ReactNode
}) => (
  <Timeline
    events={events}
    getEventKey={(event) => event.id}
    getEventTime={(event) => event.createdAt}
    group={group}
    loading={loading}
    emptyState={emptyState ?? 'No workflows yet'}
  >
    {(event) => (
      <TimelineItem
        status={event.status}
        time={event.createdAt}
        actions={<WorkflowActions event={event} />}
        title={
          <span className="flex min-w-0 flex-wrap items-center gap-2">
            <Link
              href={`/org-example/workflows/${event.id}`}
              className="w-fit"
            >
              {event.name}
            </Link>
            {event.drift ? <Badge variant="code">drift scan</Badge> : null}
          </span>
        }
      >
        <ID value={event.id} truncate />
        <Text variant="caption" color="secondary">
          Started by {event.createdBy}
          {event.commit ? ` at ${event.commit}` : ''}
        </Text>
        {event.note ? (
          <Text variant="caption" color="secondary">
            {event.note}
          </Text>
        ) : null}
      </TimelineItem>
    )}
  </Timeline>
)

export const Overview = () => (
  <ComponentDocs
    name="Timeline"
    tier="organism"
    summary="A dated list of resource history events that fills whatever width it is given."
    use={[
      'Render resource history: builds, deploys, workflows, runs, and config changes.',
      'Compose search, filters, and Pagination around Timeline, not inside it.',
      'Render each event as a TimelineItem the resource owns.',
    ]}
    avoid={[
      'Do not reserve a side column or fixed width — Timeline fills its container.',
      'Do not sort or paginate the loaded page on the client.',
      'Do not use it for uptime bars or span traces; those are their own components.',
    ]}
    rules={[
      'Events are grouped into local calendar days and labelled Today, Yesterday, or the date.',
      'Events are rendered in the order given; the API owns sorting.',
      'getEventKey is required so repeated timestamps cannot collide.',
      'Events with no usable timestamp group under Undated.',
      'Loading keeps the day heading and item rhythm as skeletons.',
    ]}
    props={[
      {
        name: 'events',
        type: 'T[]',
        description: 'Resolved events for the current API page.',
      },
      {
        name: 'children',
        type: '(event: T, index: number) => ReactNode',
        description: 'Required resource-owned event renderer.',
      },
      {
        name: 'getEventKey',
        type: '(event: T, index: number) => Key',
        description: 'Stable identity for each event.',
      },
      {
        name: 'getEventTime',
        type: '(event: T) => string | undefined',
        description: 'ISO timestamp used for day grouping.',
      },
      {
        name: 'group',
        type: 'boolean',
        default: 'true',
        description: 'Groups events under day headings.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders skeleton events in place of content.',
      },
      {
        name: 'loadingItems',
        type: 'number',
        default: '5',
        description: 'Skeleton event count while loading.',
      },
      {
        name: 'emptyState',
        type: 'ReactNode',
        default: "'No events yet'",
        description: 'Content shown when the page has no events.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <WorkflowsTimeline />
  </div>
)

export const Ungrouped = () => (
  <div className="p-8">
    <WorkflowsTimeline group={false} />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <WorkflowsTimeline loading />
  </div>
)

export const Empty = () => (
  <div className="p-8">
    <WorkflowsTimeline events={[]} />
  </div>
)

export const SingleEvent = () => (
  <div className="p-8">
    <WorkflowsTimeline events={WORKFLOWS.slice(0, 1)} />
  </div>
)

export const UndatedEvents = () => (
  <div className="p-8">
    <WorkflowsTimeline
      events={[
        ...WORKFLOWS.slice(0, 2),
        {
          id: 'wfmissingtimestamp0000001',
          name: 'Event with no timestamp',
          status: 'queued',
          createdAt: '',
          createdBy: 'engineer@example.com',
        },
      ]}
    />
  </div>
)

export const InCard = () => (
  <div className="p-8">
    <Card className="flex flex-col gap-4">
      <Text variant="heading">Install history</Text>
      <WorkflowsTimeline events={WORKFLOWS.slice(0, 4)} />
    </Card>
  </div>
)

export const InPanel = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Install history">
          <WorkflowsTimeline />
        </Panel>
      )
    }
  />
)

export const NarrowContainer = () => (
  <div className="max-w-[340px] p-4">
    <WorkflowsTimeline />
  </div>
)

export const ResponsiveContainer = () => {
  const [wide, setWide] = useState(false)

  return (
    <div className="flex flex-col gap-4 p-8">
      <Button className="self-start" onClick={() => setWide((value) => !value)}>
        {wide ? 'Make container narrow' : 'Make container wide'}
      </Button>
      <div className={wide ? 'w-full' : 'max-w-sm'}>
        <WorkflowsTimeline />
      </div>
    </div>
  )
}

export const WithListControls = () => {
  const list = useListQueryState({ pageSize: 6 })
  const filtered = WORKFLOWS.filter((event) =>
    event.name.toLowerCase().includes(list.search.toLowerCase())
  )

  return (
    <div className="flex flex-col gap-4 p-8">
      <ListSearch
        value={list.search}
        onValueChange={list.setSearch}
        placeholder="Search workflows"
        aria-label="Search workflows"
        className="w-full max-w-80"
      />
      <WorkflowsTimeline
        events={filtered}
        emptyState="No workflows match this search"
      />
      <Pagination
        offset={list.offset}
        pageSize={list.pageSize}
        hasNext={list.offset < 12}
        onOffsetChange={list.setOffset}
      />
    </div>
  )
}
