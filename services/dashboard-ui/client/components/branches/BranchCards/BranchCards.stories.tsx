export default {
  title: 'Branches/BranchCards',
}

import { BranchCards } from './BranchCards'
import { manyBranchCards, mockBranchCards } from './BranchCards.fixtures'

export const Default = () => <BranchCards cards={mockBranchCards} />

export const ManyBranches = () => <BranchCards cards={manyBranchCards} />

export const ManyBranchesExpanded = () => (
  <BranchCards cards={manyBranchCards} initialExpanded />
)

export const Loading = () => <BranchCards cards={[]} isLoading />

export const Empty = () => <BranchCards cards={[]} />

export const WithPagination = () => (
  <BranchCards
    cards={mockBranchCards}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)
