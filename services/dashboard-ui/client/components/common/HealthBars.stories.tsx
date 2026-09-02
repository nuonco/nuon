export default {
  title: 'Common/HealthBars',
}

import { HealthBars, type IHealthBar } from './HealthBars'
import { Text } from './Text'

const severityClasses = [
  'bg-green-600 dark:bg-green-500',
  'bg-lime-500 dark:bg-lime-400',
  'bg-yellow-500 dark:bg-yellow-400',
  'bg-orange-500 dark:bg-orange-400',
  'bg-red-600 dark:bg-red-500',
]

const dayBars: IHealthBar[] = Array.from({ length: 30 }, (_, i) => {
  const bucket = i % 11 === 0 ? 4 : i % 7 === 0 ? 2 : 0
  return {
    key: i,
    colorClass: severityClasses[bucket],
    ariaLabel: `Day ${i + 1}`,
    tooltip: (
      <div className="flex flex-col gap-1 w-40">
        <Text variant="subtext" weight="strong">
          {i + 1} days ago
        </Text>
        <Text variant="label" theme="neutral">
          {(100 - bucket * 5).toFixed(2)}% uptime
        </Text>
      </div>
    ),
  }
})

const runnerColors = ['bg-green-500', 'bg-red-500', 'bg-cool-grey-500']

const minuteBars: IHealthBar[] = Array.from({ length: 60 }, (_, i) => {
  const bucket = i % 17 === 0 ? 1 : i % 23 === 0 ? 2 : 0
  return {
    key: i,
    colorClass: runnerColors[bucket],
    ariaLabel: `Minute ${i}`,
    tooltip: (
      <div className="flex flex-col w-36">
        <Text variant="label" weight="strong">
          {bucket === 0 ? 'Healthy' : bucket === 2 ? 'Unknown' : 'Unhealthy'}
        </Text>
        <Text variant="subtext">{i} minutes ago</Text>
      </div>
    ),
  }
})

export const InstallDaySeverity = () => (
  <div className="max-w-xl p-4">
    <HealthBars animated grow barClassName="h-8 rounded-xs" bars={dayBars} />
  </div>
)

export const RunnerHeartbeat = () => (
  <div className="max-w-xl p-4">
    <HealthBars
      animated
      grow
      barClassName="h-8 rounded-xs"
      emptyMessage="No health data"
      bars={minuteBars}
    />
  </div>
)

export const WithLeadingRegion = () => (
  <div className="max-w-xl p-4">
    <HealthBars
      animated
      grow
      barClassName="h-8 rounded-xs"
      barWrapperClassName="max-w-[18px]"
      leading={
        <div className="flex flex-1 min-w-[8rem] items-center justify-center h-8 px-2 rounded-xs border border-dashed border-cool-grey-300 dark:border-dark-grey-600 bg-[repeating-linear-gradient(135deg,rgba(148,163,184,0.16)_0px,rgba(148,163,184,0.16)_2px,transparent_2px,transparent_7px)]">
          <Text variant="label" theme="neutral" className="truncate">
            No health data
          </Text>
        </div>
      }
      bars={dayBars.slice(0, 3)}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-xl p-4">
    <HealthBars grow emptyMessage="No health data" bars={[]} />
  </div>
)
