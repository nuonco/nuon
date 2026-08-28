import { Button } from '@/components/common/Button'
import { BranchRunApproval } from './BranchRunApproval'

export default {
  title: 'Branches/BranchRunApproval',
}

const mockActions = <Button variant="primary">Approve</Button>

export const Default = () => (
  <BranchRunApproval
    items={[
      {
        key: 'step-1',
        groupName: 'uat',
        onReview: () => {},
        actions: mockActions,
      },
    ]}
  />
)

export const MultipleGroups = () => (
  <BranchRunApproval
    items={[
      {
        key: 'step-1',
        groupName: 'uat',
        onReview: () => {},
        actions: mockActions,
      },
      {
        key: 'step-2',
        groupName: 'production',
        onReview: () => {},
        actions: mockActions,
      },
    ]}
  />
)

export const Empty = () => <BranchRunApproval items={[]} />
