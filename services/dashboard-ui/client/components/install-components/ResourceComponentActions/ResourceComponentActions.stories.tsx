export default {
  title: 'Features / Installs / Components / Resource component actions',
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
    currentDeployStatus="active"
    onViewDetails={() => {}}
  />
)

export const Disabled = () => (
  <ResourceComponentActions
    component={component}
    enabled={false}
    onViewDetails={() => {}}
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
    currentDeployStatus="active"
    variant="image"
  />
)
