export default {
  title: 'Runners/Management/DeprovisionRunner',
}

import { DeprovisionRunnerButton } from './DeprovisionRunner'

export const Default = () => (
  <div className="p-4">
    <DeprovisionRunnerButton onOpen={() => {}} />
  </div>
)

export const MenuButton = () => (
  <div className="p-4">
    <DeprovisionRunnerButton isMenuButton onOpen={() => {}} />
  </div>
)
