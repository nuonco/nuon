export default {
  title: 'Common/Group',
}

import { Group } from './Group'
import { Badge } from './Badge'
import { Button } from './Button'
import { Text } from './Text'

export const Default = () => (
  <Group>
    <Button>Save</Button>
    <Button variant="secondary">Cancel</Button>
  </Group>
)

export const WithBadges = () => (
  <Group gap={2}>
    <Badge theme="success">Active</Badge>
    <Badge theme="warning">Pending</Badge>
    <Badge theme="danger">Failed</Badge>
  </Group>
)

export const GapSizes = () => (
  <div className="flex flex-col gap-6">
    {([1, 2, 3, 4, 6, 8] as const).map((gap) => (
      <div key={gap}>
        <Text variant="subtext" className="mb-1">gap={gap}</Text>
        <Group gap={gap}>
          <Badge>One</Badge>
          <Badge>Two</Badge>
          <Badge>Three</Badge>
        </Group>
      </div>
    ))}
  </div>
)

export const JustifyBetween = () => (
  <Group justify="between">
    <Text variant="heading-3">Page title</Text>
    <Group gap={2}>
      <Button variant="secondary">Cancel</Button>
      <Button>Save</Button>
    </Group>
  </Group>
)

export const Wrap = () => (
  <div style={{ maxWidth: 300 }}>
    <Group wrap gap={2}>
      {Array.from({ length: 8 }, (_, i) => (
        <Badge key={i}>Tag {i + 1}</Badge>
      ))}
    </Group>
  </div>
)
