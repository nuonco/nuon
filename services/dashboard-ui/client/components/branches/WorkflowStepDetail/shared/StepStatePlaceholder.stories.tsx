export default {
  title: 'Branches/WorkflowStepDetail/StepStatePlaceholder',
}

import { StepStatePlaceholder } from './StepStatePlaceholder'

export const Loading = () => (
  <StepStatePlaceholder variant="loading">
    Starting component builds
  </StepStatePlaceholder>
)

export const Pending = () => (
  <StepStatePlaceholder variant="pending">
    Waiting to start component builds
  </StepStatePlaceholder>
)
