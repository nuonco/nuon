import type { Meta, StoryObj } from '@ladle/react'
import { DeploysSkeleton } from './DeploysSkeleton'

export default {
  title: 'Deploys/DeploySwitcher/DeploysSkeleton',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <div className="flex flex-col gap-2 w-64">
      <DeploysSkeleton />
    </div>
  ),
}

export const CustomLimit: StoryObj = {
  render: () => (
    <div className="flex flex-col gap-2 w-64">
      <DeploysSkeleton limit={3} />
    </div>
  ),
}
