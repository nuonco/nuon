export default {
  title: 'Common/Stack',
}

import { Stack } from './Stack'
import { Badge } from './Badge'
import { Text } from './Text'

export const Default = () => (
  <Stack>
    <Text variant="h3">Section title</Text>
    <Text>First paragraph of content.</Text>
    <Text>Second paragraph of content.</Text>
  </Stack>
)

export const GapSizes = () => (
  <div className="flex gap-8">
    {([1, 2, 3, 4, 6, 8] as const).map((gap) => (
      <div key={gap}>
        <Text variant="subtext" className="mb-1">gap={gap}</Text>
        <Stack gap={gap}>
          <Badge>One</Badge>
          <Badge>Two</Badge>
          <Badge>Three</Badge>
        </Stack>
      </div>
    ))}
  </div>
)

export const FormLayout = () => (
  <Stack gap={3} style={{ maxWidth: 320 }}>
    <Stack gap={1}>
      <Text variant="subtext">Name</Text>
      <input className="border rounded px-2 py-1" placeholder="Enter name" />
    </Stack>
    <Stack gap={1}>
      <Text variant="subtext">Email</Text>
      <input className="border rounded px-2 py-1" placeholder="Enter email" />
    </Stack>
    <Stack gap={1}>
      <Text variant="subtext">Description</Text>
      <textarea className="border rounded px-2 py-1" rows={3} placeholder="Enter description" />
    </Stack>
  </Stack>
)
