import { DateTime } from 'luxon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Badge } from '../atoms/Badge'
import { Button } from '../atoms/Button'
import { Icon } from '../atoms/Icon'
import { Link } from '../atoms/Link'
import { Text } from '../atoms/Text'
import { ID } from './ID'
import { TimelineItem } from './TimelineItem'

export default {
  title: 'lite/molecules/TimelineItem',
}

const List = ({ children }: { children: React.ReactNode }) => (
  <ol className="flex min-w-0 flex-col">{children}</ol>
)

export const Overview = () => (
  <ComponentDocs
    name="TimelineItem"
    tier="molecule"
    summary="One event in a Timeline: status marker, title, timestamp, actions, and resource-owned body content."
    use={[
      'Render a single history event inside a Timeline.',
      'Put the resource link in title so the name is what navigates.',
      'Put IDs, commits, authors, and durations in children.',
    ]}
    avoid={[
      'Do not use it outside an ordered list — it renders an li.',
      'Do not add a second metadata column; the item is one readable column.',
      'Do not make the whole item clickable.',
    ]}
    rules={[
      'The connector line is hidden on the last item of its list.',
      'time takes an ISO string and renders relative, with the absolute time on hover.',
      'An item with no status renders a neutral marker rather than shifting the layout.',
      'loading keeps the marker and spacing and replaces text with skeletons.',
    ]}
    props={[
      {
        name: 'title',
        type: 'ReactNode',
        description: 'Event heading, usually the resource link.',
      },
      {
        name: 'status',
        type: 'string | TCompositeStatus',
        description: 'API status mapped to the marker theme and icon.',
      },
      {
        name: 'statusIcon',
        type: 'TIconVariant',
        description:
          'Overrides the marker icon without changing the status theme.',
      },
      {
        name: 'time',
        type: 'string',
        description: 'ISO timestamp rendered as relative time.',
      },
      {
        name: 'actions',
        type: 'ReactNode',
        description: 'Controls aligned to the end of the title row.',
      },
      {
        name: 'children',
        type: 'ReactNode',
        description: 'Resource-owned body content below the timestamp.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the item as a skeleton.',
      },
    ]}
  />
)

export const Title = () => (
  <div className="p-8">
    <List>
      <TimelineItem title="Payments API build" />
      <TimelineItem
        title={
          <Link href="/org-example/builds/bld-1" className="w-fit">
            Payments API build
          </Link>
        }
      />
    </List>
  </div>
)

export const Status = () => (
  <div className="p-8">
    <List>
      <TimelineItem status="active" title="Deploy succeeded" />
      <TimelineItem status="in-progress" title="Deploy running" />
      <TimelineItem status="failed" title="Deploy failed" />
      <TimelineItem
        status="cancelled"
        statusIcon="ProhibitIcon"
        title="Deploy cancelled"
      />
      <TimelineItem status="queued" title="Deploy queued" />
      <TimelineItem title="Config synced without a status" />
    </List>
  </div>
)

export const Time = () => (
  <div className="p-8">
    <List>
      <TimelineItem
        status="active"
        title="Two minutes ago"
        time={DateTime.now().minus({ minutes: 2 }).toISO()}
      />
      <TimelineItem
        status="active"
        title="Six hours ago"
        time={DateTime.now().minus({ hours: 6 }).toISO()}
      />
      <TimelineItem
        status="active"
        title="Three days ago"
        time={DateTime.now().minus({ days: 3 }).toISO()}
      />
      <TimelineItem status="active" title="No timestamp provided" />
    </List>
  </div>
)

export const Actions = () => (
  <div className="p-8">
    <List>
      <TimelineItem
        status="in-progress"
        title="Payments API deploy"
        time={DateTime.now().minus({ seconds: 90 }).toISO()}
        actions={
          <Button size="sm" variant="ghost">
            Cancel
          </Button>
        }
      />
      <TimelineItem
        status="active"
        title="Dashboard deploy"
        time={DateTime.now().minus({ minutes: 45 }).toISO()}
        actions={
          <Button
            size="sm"
            variant="ghost"
            iconOnly
            aria-label="Deploy actions"
            tooltip="Deploy actions"
          >
            <Icon variant="DotsThreeIcon" size={16} aria-hidden />
          </Button>
        }
      />
    </List>
  </div>
)

export const BodyContent = () => (
  <div className="p-8">
    <List>
      <TimelineItem
        status="active"
        title={
          <Link href="/org-example/builds/bld-1" className="w-fit">
            Payments API build
          </Link>
        }
        time={DateTime.now().minus({ minutes: 32 }).toISO()}
      >
        <ID value="bldq7fplr1up5atx5zpxotbabm" truncate />
        <Text variant="caption" color="secondary">
          Built by engineer@example.com
        </Text>
      </TimelineItem>
      <TimelineItem
        status="failed"
        title={
          <span className="flex min-w-0 flex-wrap items-center gap-2">
            <Link href="/org-example/deploys/dep-2" className="w-fit">
              Dashboard deploy
            </Link>
            <Badge variant="code">drift scan</Badge>
          </span>
        }
        time={DateTime.now().minus({ hours: 4 }).toISO()}
      >
        <ID value="depq7fplr1up5atx5zpxotbabm" truncate />
        <Text variant="caption" color="secondary">
          Terraform apply exited with status 1
        </Text>
      </TimelineItem>
    </List>
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <List>
      <TimelineItem title="" loading />
      <TimelineItem title="" loading />
      <TimelineItem title="" loading />
    </List>
  </div>
)

export const LongContent = () => (
  <div className="p-8">
    <List>
      <TimelineItem
        status="failed"
        title="A deploy of a component whose name runs well past the width available to the timeline item"
        time={DateTime.now().minus({ minutes: 8 }).toISO()}
        actions={
          <Button size="sm" variant="ghost">
            Retry
          </Button>
        }
      >
        <ID value="depq7fplr1up5atx5zpxotbabmq7fplr1up5atx5zpxotbabm" />
        <Text variant="caption" color="secondary">
          The runner could not reach the cluster endpoint and retried three
          times before giving up on this component.
        </Text>
      </TimelineItem>
      <TimelineItem status="active" title="Short follow-up event" />
    </List>
  </div>
)

export const NarrowContainer = () => (
  <div className="max-w-[340px] p-4">
    <List>
      <TimelineItem
        status="in-progress"
        title={
          <Link href="/org-example/deploys/dep-1" className="w-fit">
            Payments API deploy
          </Link>
        }
        time={DateTime.now().minus({ minutes: 3 }).toISO()}
        actions={
          <Button size="sm" variant="ghost">
            Cancel
          </Button>
        }
      >
        <ID value="depq7fplr1up5atx5zpxotbabm" truncate />
      </TimelineItem>
      <TimelineItem
        status="active"
        title="Sandbox reprovisioned"
        time={DateTime.now().minus({ hours: 26 }).toISO()}
      >
        <Text variant="caption" color="secondary">
          Triggered by a scheduled drift run.
        </Text>
      </TimelineItem>
    </List>
  </div>
)
