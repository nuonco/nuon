import { Skeleton } from '@/components/common/Skeleton'

export const LogLineSkeleton = () => {
  return (
    <div className="grid grid-cols-[3rem_15rem_3rem_1fr] gap-6 py-2">
      <Skeleton height="17px" width="40px" />
      <Skeleton height="17px" width="240px" />
      <Skeleton height="17px" width="50px" />
      <Skeleton height="17px" width="100%" />
    </div>
  )
}
