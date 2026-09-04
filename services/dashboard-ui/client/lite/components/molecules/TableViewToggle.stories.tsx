import { useState } from 'react'
import type { TTableView } from '../../hooks/use-table-view'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { TableViewToggle } from './TableViewToggle'

export default {
  title: 'lite/molecules/TableViewToggle',
}

export const Overview = () => (
  <ComponentDocs
    name="TableViewToggle"
    tier="molecule"
    summary="A controlled ToggleButton configured for table and card-grid presentations."
    use={[
      'Let Table render this control when both presentations are available.',
      'Use the shared session preference as its controlled value.',
    ]}
    avoid={[
      'Do not show the toggle while narrow space forces cards.',
      'Do not use it to change the loaded rows or query state.',
    ]}
    rules={[
      'Both choices expose pressed state and accessible labels.',
      'Changing presentation does not change collection state.',
    ]}
    props={[
      {
        name: 'value',
        type: "'table' | 'cards'",
        description: 'The preferred collection presentation.',
      },
      {
        name: 'onValueChange',
        type: "(value: 'table' | 'cards') => void",
        description: 'Receives the next preferred presentation.',
      },
    ]}
  />
)

export const Interactive = () => {
  const [value, setValue] = useState<TTableView>('table')

  return (
    <div className="p-8">
      <TableViewToggle value={value} onValueChange={setValue} />
    </div>
  )
}

export const CardsSelected = () => (
  <div className="p-8">
    <TableViewToggle value="cards" onValueChange={() => {}} />
  </div>
)
