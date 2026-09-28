export default {
  title: 'Features / Deploys / Deploy switcher',
}

import { DeploySwitcher } from './DeploySwitcher'

export const Default = () => (
  <DeploySwitcher
    componentId="comp-1"
    deployId="deploy-1"
  />
)
