import { Card, type ICard } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'

export const RunnerDetailsCardSkeleton = (props: Omit<ICard, 'children'>) => {
  return (
    <Card {...props}>
      <Skeleton height="24px" width="106px" />

      <div className="grid gap-6 md:grid-cols-2">
        <LabeledValue label={<Skeleton height="17px" width="34px" />}>
          <Skeleton height="23px" width="75px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="68px" />}>
          <Skeleton height="23px" width="110px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="41px" />}>
          <Skeleton height="23px" width="50px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="45px" />}>
          <Skeleton height="23px" width="54px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="148px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="215px" />
        </LabeledValue>
      </div>
    </Card>
  )
}
