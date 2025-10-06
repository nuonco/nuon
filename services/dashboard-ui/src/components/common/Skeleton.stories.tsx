import { Skeleton } from './Skeleton'

export const Default = () => <Skeleton />

export const MultipleLines = () => <Skeleton lines={3} />

export const CustomDimensions = () => (
  <Skeleton lines={2} width={['50%', '75%']} height="2rem" />
)
