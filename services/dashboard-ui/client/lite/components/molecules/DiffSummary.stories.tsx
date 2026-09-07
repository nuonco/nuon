import { emptyDiffSummary } from '../../lib/diffs'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { DiffSummary } from './DiffSummary'

export default {
  title: 'lite/molecules/DiffSummary',
}

export const Overview = () => (
  <ComponentDocs
    name="DiffSummary"
    tier="molecule"
    summary="Provider-neutral counts for the operations in a plan."
    use={[
      'Summarize a complete plan before filters are applied.',
      'Show the same operation vocabulary for every provider.',
    ]}
    avoid={[
      'Do not calculate counts from filtered sections.',
      'Do not use provider-specific action names.',
    ]}
    rules={[
      'Providers choose which operations are relevant, including zero counts.',
      'Create, update and delete use the same color tokens as diff sections.',
    ]}
    props={[
      {
        name: 'summary',
        type: 'IPlanDiffSummary',
        description: 'Counts for every normalized operation. Required.',
      },
      {
        name: 'operations',
        type: 'TDiffOperation[]',
        default: 'all operations',
        description: 'Operations displayed by this provider.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <DiffSummary
      summary={{
        ...emptyDiffSummary(),
        create: 3,
        update: 2,
        delete: 1,
      }}
      operations={['create', 'update', 'delete']}
    />
  </div>
)

export const AllZero = () => (
  <div className="p-8">
    <DiffSummary
      summary={emptyDiffSummary()}
      operations={['create', 'update', 'delete']}
    />
  </div>
)
