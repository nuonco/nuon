import type { Meta, StoryObj } from '@ladle/react'
import { RenderedOutputs } from './RenderedOutputs'

export default {
  title: 'Deploys/RenderedOutputs',
} satisfies Meta

export const Default: StoryObj = {
  render: () => <RenderedOutputs />,
}
