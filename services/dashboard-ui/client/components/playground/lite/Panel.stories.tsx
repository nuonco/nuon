import { Panel } from './Panel'
import { PlaceholderGrid } from './PlaceholderGrid'

export default {
  title: 'Playground/Lite/Panel',
}

export const Default = () => (
  <Panel title="Recent activity" action="View all">
    <PlaceholderGrid rows={4} height="h-[2rem]" />
  </Panel>
)

export const NoAction = () => (
  <Panel title="Usage">
    <PlaceholderGrid rows={2} height="h-[3rem]" />
  </Panel>
)
