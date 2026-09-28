export default {
  title: 'UI / Dependency view toggle',
}

import { useState } from 'react'
import {
  DependencyViewToggle,
  type TDependencyViewMode,
} from './DependencyViewToggle'

export const Default = () => {
  const [value, setValue] = useState<TDependencyViewMode>('graph')
  return <DependencyViewToggle value={value} onChange={setValue} />
}

export const TableSelected = () => {
  const [value, setValue] = useState<TDependencyViewMode>('table')
  return <DependencyViewToggle value={value} onChange={setValue} />
}
