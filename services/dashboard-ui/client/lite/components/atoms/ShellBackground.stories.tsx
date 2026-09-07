import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Card } from './Card'
import { ShellBackground } from './ShellBackground'
import { Text } from './Text'

export default {
  title: 'lite/atoms/ShellBackground',
}

export const Overview = () => (
  <ComponentDocs
    name="ShellBackground"
    tier="atom"
    summary="A subtle ambient dot grid shared by Lite application shells."
    use={[
      'Render once behind DashboardShell or FocusShell content.',
      'Keep graph-aware grids inside their graph viewport instead.',
    ]}
    avoid={[
      'Do not use it as a React Flow background that needs to pan or zoom.',
      'Do not place interactive content inside the decorative layer.',
    ]}
    rules={[
      'The grid is decorative, ignores pointer input, and is hidden from assistive technology.',
      'Theme tokens keep the dots quieter than foreground content.',
      'The vertical fade prevents a hard patterned edge against the viewport.',
    ]}
  />
)

export const Default = () => (
  <div className="relative isolate h-96 overflow-hidden bg-surface-default p-12">
    <ShellBackground />
    <div className="relative z-10 mx-auto max-w-md">
      <Card shadow="floating">
        <Text variant="heading">Graph-ready workspace</Text>
        <Text as="p" variant="caption" color="secondary">
          Shell content remains legible over the ambient dot matrix.
        </Text>
      </Card>
    </div>
  </div>
)
