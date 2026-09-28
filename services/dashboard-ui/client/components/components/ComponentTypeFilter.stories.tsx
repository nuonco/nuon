export default {
  title: 'Features / Components / Type filter',
}

import { ComponentTypeFilterDropdown } from './ComponentTypeFilter'

export const Dropdown = () => <ComponentTypeFilterDropdown />

export const Inline = () => <ComponentTypeFilterDropdown isNotDropdown />
