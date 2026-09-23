export default {
  title: 'Install Components/ResourceComponentActions',
}

import type { TComponent } from '@/types'
import { ResourceComponentActions } from './ResourceComponentActions'

const component = {
  id: 'cmp-1',
  name: 'api',
  type: 'helm_chart',
  app_id: 'app-1',
} as TComponent

export const Component = () => (
  <ResourceComponentActions
    component={component}
    currentBuildId="bld-1"
    currentDeployStatus="active"
  />
)

export const Image = () => (
  <ResourceComponentActions
    component={{
      ...component,
      id: 'cmp-image-1',
      name: 'api-image',
      type: 'external_image',
    }}
    currentBuildId="bld-image-1"
    currentDeployStatus="active"
    variant="image"
  />
)
