import type { FC } from 'react'
import { Button, Text } from '@/stratus/components'
import { IPageProps } from '@/types'

const StratusDasboard: FC<IPageProps<'org-id'>> = () => {
  return (
    <div className="flex flex-col gap-4 p-4">
      <Text variant="h1" weight="stronger">
        Buttons
      </Text>
      <div className="flex gap-4">
        <div className="flex gap-4">
          <Button variant="primary">Primary</Button>
          <Button>Secondary</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="danger">Danger</Button>
        </div>
      </div>

      <Text variant="h1" weight="stronger">
        Typography
      </Text>
      <div className="flex gap-4">
        <div className="flex flex-col">
          <Text variant="h1">
            H1: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="h2">
            h2: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="h3">
            h3: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="base">
            base: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="body">
            body: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="subtext">
            subtext: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="label">
            label: The quick brown fox jumps over the lazy dog.
          </Text>
        </div>
        <div className="flex flex-col">
          <Text family="mono" variant="h1">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="h2">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="base">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="body">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="subtext">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="label">
            The quick brown fox jumps over the lazy dog.
          </Text>
        </div>
      </div>
    </div>
  )
}

export default StratusDasboard
