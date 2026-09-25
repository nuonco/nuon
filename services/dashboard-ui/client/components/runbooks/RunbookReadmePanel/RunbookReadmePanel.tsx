import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'

export interface IRunbookReadmePanel extends Omit<IPanel, 'heading'> {
  readme?: string
  runbookName?: string
}

export const RunbookReadmePanel = ({
  readme,
  runbookName,
  ...props
}: IRunbookReadmePanel) => (
  <Panel
    heading={runbookName ? `${runbookName} readme` : 'Readme'}
    size="half"
    {...props}
  >
    {readme ? (
      <Markdown content={readme} mode="install" />
    ) : (
      <Text theme="neutral">No readme configured.</Text>
    )}
  </Panel>
)
