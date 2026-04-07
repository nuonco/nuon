import type { Meta, StoryObj } from '@ladle/react'
import { Logo } from './Logo'

export default {
  title: 'Common/Logo',
} satisfies Meta

export const System: StoryObj = {
  render: () => <Logo />,
}

export const Light: StoryObj = {
  render: () => (
    <div className="bg-dark-grey-900 p-4">
      <Logo variant="light" />
    </div>
  ),
}

export const Dark: StoryObj = {
  render: () => (
    <div className="bg-white p-4">
      <Logo variant="dark" />
    </div>
  ),
}
