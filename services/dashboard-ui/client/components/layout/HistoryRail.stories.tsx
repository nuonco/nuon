export default {
  title: 'Layout/HistoryRail',
}

import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { HistoryPanelButton, HistoryRail } from './HistoryRail'
import { PageSection } from './PageSection'

const history = (
  <div className="flex flex-col gap-3">
    {['Ran 2 hours ago', 'Ran 1 day ago', 'Ran 3 days ago'].map((entry) => (
      <Card key={entry} className="!p-4 !gap-2">
        <Text variant="subtext">{entry}</Text>
      </Card>
    ))}
  </div>
)

const sections = (
  <>
    <Card>
      <Text variant="base" weight="strong">
        Configuration
      </Text>
      <Text theme="neutral">Section body</Text>
    </Card>
    <Card>
      <Text variant="base" weight="strong">
        Steps
      </Text>
      <Text theme="neutral">Section body</Text>
    </Card>
  </>
)

export const Default = () => (
  <PageSection className="@container">
    <HistoryRail title="Run history" history={history}>
      {sections}
    </HistoryRail>
  </PageSection>
)

export const Narrow = () => (
  <div className="@container max-w-2xl">
    <PageSection>
      <div className="flex justify-end">
        <HistoryPanelButton title="Run history" history={history} />
      </div>
      <HistoryRail title="Run history" history={history}>
        {sections}
      </HistoryRail>
    </PageSection>
  </div>
)
