import type { ReactNode } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Text } from '../atoms/Text'
import { GraphNodeCard } from './GraphNodeCard'

export default {
  title: 'lite/molecules/GraphNodeCard',
}

export const Overview = () => (
  <ComponentDocs
    name="GraphNodeCard"
    tier="molecule"
    summary="Shared chrome for a graph node: a card, selected and hover treatment, a focus ring, and a status accent."
    use={[
      'Compose every resource node type on Graph from this card so nodes share one hover, selected, and focus treatment.',
      'Pass href so the node is a Link to a panel. Omit href for a node that cannot be opened.',
    ]}
    avoid={[
      'Do not use it for a container. Containers are a labelled translucent region, not a GraphNodeCard.',
      'Do not draw a custom card for a graph node. That is the drift this molecule exists to prevent.',
      'Do not wrap the node in a button or attach onClick. Pass href so the node is a Link.',
      'Do not use it outside a graph. Grouped content on a page is Card.',
    ]}
    rules={[
      'The caller owns the node size. GraphNodeCard fills the box it is given.',
      'A node with no href is not interactive: no hover tint, no pointer cursor, no focus ring.',
      'Status colour comes through Status. Do not tint the card by hand.',
      'Selected is an accent outline, not a colour-only fill.',
      'Put the node’s identity in children. GraphNodeCard is chrome, not a resource renderer.',
    ]}
    props={[
      {
        name: 'href',
        type: 'string',
        description:
          'Panel or page destination. Renders a Link with hover and a focus ring. Omit for a non-interactive node.',
      },
      {
        name: 'selected',
        type: 'boolean',
        default: 'false',
        description:
          'Marks the node whose panel is open. Draws an accent outline and sets aria-current on the link.',
      },
      {
        name: 'status',
        type: 'string | TCompositeStatus',
        description:
          'API status rendered as a Status icon accent. Colour is never the only signal.',
      },
    ]}
  />
)

const Sample = ({
  name,
  detail,
}: {
  name: string
  detail: string
}) => (
  <>
    <Text weight="medium">{name}</Text>
    <Text variant="caption" color="tertiary">
      {detail}
    </Text>
  </>
)

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="flex flex-wrap items-start gap-4 p-8">{children}</div>
)

const NODE = 'h-20 w-52'

export const Default = () => (
  <Frame>
    <GraphNodeCard href="?panel=component:cmp_api" className={NODE}>
      <Sample name="payments-api" detail="Helm chart" />
    </GraphNodeCard>
  </Frame>
)

export const Selected = () => (
  <Frame>
    <GraphNodeCard href="?panel=component:cmp_api" className={NODE}>
      <Sample name="payments-api" detail="Helm chart" />
    </GraphNodeCard>
    <GraphNodeCard
      href="?panel=component:cmp_db"
      selected
      status="active"
      className={NODE}
    >
      <Sample name="payments-db" detail="Terraform module" />
    </GraphNodeCard>
  </Frame>
)

export const Statuses = () => (
  <Frame>
    <GraphNodeCard
      href="?panel=component:cmp_ok"
      status="healthy"
      className={NODE}
    >
      <Sample name="payments-api" detail="Helm chart" />
    </GraphNodeCard>
    <GraphNodeCard
      href="?panel=component:cmp_run"
      status="deploying"
      className={NODE}
    >
      <Sample name="payments-worker" detail="Kubernetes manifest" />
    </GraphNodeCard>
    <GraphNodeCard
      href="?panel=component:cmp_fail"
      status="failed"
      className={NODE}
    >
      <Sample name="payments-db" detail="Terraform module" />
    </GraphNodeCard>
  </Frame>
)

export const NotInteractive = () => (
  <Frame>
    <GraphNodeCard status="pending" className={NODE}>
      <Sample name="sandbox" detail="No panel" />
    </GraphNodeCard>
  </Frame>
)

export const Keyboard = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      Tab between the linked nodes. Each focus ring is on the link, not a
      hand-rolled click target.
    </Text>
    <div className="flex flex-wrap items-start gap-4">
      <GraphNodeCard href="?panel=component:cmp_api" className={NODE}>
        <Sample name="payments-api" detail="Helm chart" />
      </GraphNodeCard>
      <GraphNodeCard
        href="?panel=component:cmp_db"
        status="healthy"
        className={NODE}
      >
        <Sample name="payments-db" detail="Terraform module" />
      </GraphNodeCard>
      <GraphNodeCard status="pending" className={NODE}>
        <Sample name="sandbox" detail="Skipped by tab" />
      </GraphNodeCard>
    </div>
  </div>
)
