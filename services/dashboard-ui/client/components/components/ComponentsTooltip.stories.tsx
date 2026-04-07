export default {
  title: 'Components/ComponentsTooltip',
}

import { ComponentsTooltip } from './ComponentsTooltip'
import { Text } from '@/components/common/Text'

const mockSummaries = [
  {
    id: 'comp-1',
    href: '/org-1/installs/inst-1/components/comp-1',
    title: 'api-server',
    subtitle: 'Installed',
  },
  {
    id: 'comp-2',
    href: '/org-1/installs/inst-1/components/comp-2',
    title: 'worker',
    subtitle: 'Deploying',
  },
  {
    id: 'comp-3',
    href: '/org-1/installs/inst-1/components/comp-3',
    title: 'database',
    subtitle: 'Installed',
  },
]

export const Default = () => (
  <ComponentsTooltip
    componentSummaries={mockSummaries}
    title="Components"
  >
    <Text>3 components</Text>
  </ComponentsTooltip>
)

export const Empty = () => (
  <ComponentsTooltip
    componentSummaries={[]}
    title="Components"
  >
    <Text>No components</Text>
  </ComponentsTooltip>
)
