import { useState } from 'react'
import type { TDiffOperation } from '../../lib/diffs'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { DiffFilter } from './DiffFilter'

export default {
  title: 'lite/organisms/DiffFilter',
}

const operations: TDiffOperation[] = ['create', 'update', 'delete']

const Example = () => {
  const [searchValue, setSearchValue] = useState('')
  const [selectedOperations, setSelectedOperations] = useState(
    new Set(operations)
  )

  const reset = () => {
    setSearchValue('')
    setSelectedOperations(new Set(operations))
  }

  return (
    <DiffFilter
      title="changes"
      operations={operations}
      selectedOperations={selectedOperations}
      selectedCount={7}
      totalCount={9}
      searchValue={searchValue}
      searchPlaceholder="Search by release, resource, or type"
      onSearchChange={setSearchValue}
      onOperationToggle={(operation) =>
        setSelectedOperations((selected) => {
          const next = new Set(selected)
          if (next.has(operation)) next.delete(operation)
          else next.add(operation)
          return next
        })
      }
      onOperationOnly={(operation) =>
        setSelectedOperations((selected) =>
          selected.size === 1 && selected.has(operation)
            ? new Set(operations)
            : new Set([operation])
        )
      }
      onReset={reset}
    />
  )
}

export const Overview = () => (
  <ComponentDocs
    name="DiffFilter"
    tier="organism"
    summary="Shared metadata search and operation filters for plan diffs."
    use={[
      'Filter normalized provider sections without changing their summary.',
      'Keep search and operation behavior identical across providers.',
    ]}
    avoid={[
      'Do not put expand, split, wrap or back-to-top controls here.',
      'Do not pass raw provider actions.',
    ]}
    rules={[
      'The checkbox toggles an operation; the rest of the row isolates it.',
      'An isolated row resets every operation when clicked again.',
      'Reset restores every provider operation and clears search.',
    ]}
    props={[
      {
        name: 'operations',
        type: 'TDiffOperation[]',
        description: 'Normalized operations available in this provider.',
      },
      {
        name: 'selectedOperations',
        type: 'Set<TDiffOperation>',
        description: 'Currently visible operations.',
      },
      {
        name: 'searchValue',
        type: 'string',
        description: 'Controlled metadata query.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-5xl p-8">
    <Example />
  </div>
)
