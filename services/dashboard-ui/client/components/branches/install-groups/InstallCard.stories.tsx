import type { Meta, StoryObj } from '@ladle/react'
import { InstallCard } from './InstallCard'

export default {
  title: 'Branches/InstallGroups/InstallCard',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <InstallCard
      install={{ id: 'inst-1', name: 'Production US East' } as any}
    />
  ),
}

export const Disabled: StoryObj = {
  render: () => (
    <InstallCard
      install={{ id: 'inst-2', name: 'Staging' } as any}
      isDisabled
    />
  ),
}
