import { Skeleton } from '@/components/common/Skeleton'

export const DeploysSkeleton = ({ limit = 5 }: { limit?: number }) => {
  return Array.from({ length: limit }).map((_, idx) => (
    <span
      key={`deploy-skeleton-${idx}`}
      className="flex flex-col w-full gap-1 rounded-lg border p-2"
    >
      <span className="flex items-center justify-between">
        <Skeleton height="17px" width="160px" />
        <Skeleton height="14px" width="50px" />
      </span>
      <span className="flex items-center gap-4 w-full">
        <Skeleton height="17px" width="50px" />
        <Skeleton height="14px" width="70px" />
      </span>
    </span>
  ))
}
