import { Text } from '@/components/common/Text'

export interface IStatsCard {
  label: string
  value: number | string
}

export const StatsCard = ({ label, value }: IStatsCard) => {
  return (
    <div className="flex flex-col gap-4 p-4">
      <Text variant="body" weight="strong" theme="neutral">
        {label}
      </Text>
      <Text variant="h2" weight="strong">
        {value}
      </Text>
    </div>
  )
}

export const StatsGrid = ({ stats }: { stats: IStatsCard[] }) => {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 rounded-lg border divide-x divide-y lg:divide-y-0">
      {stats.map((stat, index) => (
        <StatsCard key={index} {...stat} />
      ))}
    </div>
  )
}

