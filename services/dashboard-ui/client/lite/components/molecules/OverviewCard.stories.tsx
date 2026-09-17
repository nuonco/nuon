import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Status } from '../atoms/Status'
import { Text } from '../atoms/Text'
import { OverviewCard, OverviewCardGrid } from './OverviewCard'

export default {
  title: 'lite/molecules/OverviewCard',
}

export const Overview = () => (
  <ComponentDocs
    name="OverviewCard"
    tier="molecule"
    summary="A compact summary card for the most important facts on an entity overview."
    use={[
      'Place a short label, one primary value, and supporting context at the top of an overview page.',
      'Compose cards inside OverviewCardGrid so they stack before expanding to three or four columns.',
    ]}
    avoid={[
      'Do not use overview cards for resource collections or detailed configuration.',
      'Do not put more than four cards in one overview grid.',
      'Do not nest another Card inside an overview card.',
    ]}
    rules={[
      'The grid is one column when narrow, two at medium widths, and at most four when wide.',
      'The primary value remains scannable without reading the supporting context.',
    ]}
    props={[
      {
        name: 'title',
        type: 'string',
        description: 'Short sentence-case label for the summarized fact.',
      },
      {
        name: 'footer',
        type: 'ReactNode',
        description: 'Optional supporting metadata anchored below the main value.',
      },
      {
        name: 'columns',
        type: '3 | 4',
        default: '4',
        description: 'Maximum number of columns in OverviewCardGrid.',
      },
    ]}
  />
)

export const ThreeCards = () => (
  <div className="p-8">
    <OverviewCardGrid columns={3}>
      <OverviewCard title="Branch info">
        <Text family="mono" weight="medium">
          main
        </Text>
        <Text variant="caption" color="tertiary">
          Config v14
        </Text>
      </OverviewCard>
      <OverviewCard title="Last update">
        <Text family="mono" weight="medium">
          a1b2c3d
        </Text>
        <Text variant="caption" color="secondary">
          Update component versions
        </Text>
      </OverviewCard>
      <OverviewCard title="Installs">
        <Text variant="title">12</Text>
        <Text variant="caption" color="tertiary">
          Assigned to this branch
        </Text>
      </OverviewCard>
    </OverviewCardGrid>
  </div>
)

export const FourCards = () => (
  <div className="p-8">
    <OverviewCardGrid>
      <OverviewCard title="Health">
        <Status status="healthy" />
      </OverviewCard>
      <OverviewCard title="Drift">
        <Status status="no-drift" />
      </OverviewCard>
      <OverviewCard title="Branch">
        <Text family="mono" weight="medium">
          main
        </Text>
      </OverviewCard>
      <OverviewCard title="Last update">
        <Text family="mono" weight="medium">
          a1b2c3d
        </Text>
      </OverviewCard>
    </OverviewCardGrid>
  </div>
)
