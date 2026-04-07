import type { Meta, StoryObj } from '@ladle/react'
import { DeploySwitcher } from './DeploySwitcher'

export default {
  title: 'Deploys/DeploySwitcher',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <DeploySwitcher
      componentId="comp-1"
      deployId="deploy-1"
    />
  ),
}
