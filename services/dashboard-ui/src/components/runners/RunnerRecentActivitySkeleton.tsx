import { Skeleton } from '@/components/common/Skeleton'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'

export const RunnerRecentActivitySkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="110px" />
      <TimelineSkeleton eventCount={10} />
    </>
  )
}
