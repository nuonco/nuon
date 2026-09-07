import { ComponentDocs } from '../__stories__/ComponentDocs'
import { RouteScaffold } from './RouteScaffold'

export default {
  title: 'lite/organisms/RouteScaffold',
}

export const Overview = () => (
  <ComponentDocs
    name="RouteScaffold"
    tier="organism"
    summary="Temporary route identity content used while Lite pages are built."
    use={[
      'Use only to prove a real route and shell composition before its page is implemented.',
    ]}
    avoid={[
      'Do not treat this as the future PageTemplate.',
      'Do not add page-specific actions, data fetching, or empty states.',
    ]}
    rules={['Replace each scaffold when its real page implementation begins.']}
  />
)

export const Default = () => (
  <div className="p-4">
    <RouteScaffold
      title="Applications"
      description="Manage applications for this organization."
    />
  </div>
)
