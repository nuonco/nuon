import { Card, type ICard } from '@/components/common/Card'
import { Skeleton } from '@/components/common/Skeleton'
import { cn } from '@/utils/classnames'

export const RunnerHealthCardSkeleton = ({
  className,
  ...props
}: Omit<ICard, 'children'>) => {
  return (
    <Card className={cn('flex-auto justify-between', className)} {...props}>
      <Skeleton height="24px" width="98px" />

      <div className="flex flex-col gap-6 w-full">
        <Skeleton height="24px" width="180px" />

        <Skeleton height="56px" width="100%" />

        <Skeleton height="25px" width="100%" />
      </div>
    </Card>
  )
}
