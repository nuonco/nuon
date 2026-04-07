export default {
  title: 'Components/ComponentTypeFilter',
}

import { ComponentTypeFilterDropdown } from './ComponentTypeFilter'

export const Dropdown = () => <ComponentTypeFilterDropdown />

export const Inline = () => <ComponentTypeFilterDropdown isNotDropdown />
