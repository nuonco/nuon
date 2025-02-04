'use client'

import {
  BarElement,
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import classNames from 'classnames'
import { DateTime } from 'luxon'
import { type FC } from 'react'
import { Bar } from 'react-chartjs-2'
import { ToolTip } from '@/components/ToolTip'
import { Text } from '@/components/Typography'
import type { TRunnerRecentHeartbeat } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend
)

function getMinutesOfCurrentHour() {
  const now = new Date()
  const currentHour = now.getHours()
  const minutesArray = []

  for (let minute = 0; minute < 60; minute++) {
    minutesArray.push(
      new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate(),
        currentHour,
        minute
      )
    )
  }

  return minutesArray
}

export const RunnerHeartbeatChart: FC = () => {
  // Sample data for service availability
  // 1 means the service was available, 0 means it was unavailable
  const uptimeData = [
    1, 1, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
  ] // 60 data points for 60 minutes

  const labels = getMinutesOfCurrentHour()

  return (
    <div className="flex flex-col gap-6 w-full">
      <div className="flex items-center gap-0.5">
        {uptimeData?.map((ut, i) => (
          <ToolTip
            key={`${ut}-${i}`}
            alignment={i <= 9 ? 'left' : i >= 49 ? 'right' : 'center'}
            parentClassName="flex-auto heartbeat-item-parent"
            tipContent={
              ut === 1 ? (
                <>
                  <Text variant="med-12">Available</Text>
                  <Text variant="reg-12">{labels[i].toLocaleString()}</Text>
                </>
              ) : (
                <>
                  <Text variant="med-12">Unavailable</Text>
                  <Text variant="reg-12">{labels[i].toLocaleString()}</Text>
                </>
              )
            }
            isIconHidden
          >
            <div
              className={classNames(
                'flex-auto max-w-[16px] h-[46px] border rounded-sm heartbeat-item',
                {
                  'border-green-600 bg-green-500': ut === 1,
                  'border-red-600 bg-red-500': ut === 0,
                }
              )}
            />
          </ToolTip>
        ))}
      </div>
      <div className="flex items-center justify-between bg-black/5 dark:bg-white/5 px-4 py-1">
        {labels
          ?.filter((_, i) => (i + 1) % 12 === 0)
          ?.map((label) => (
            <Text key={label.toSring()} className="rotate-0" variant="med-12">
              {label.toLocaleString([], { hour: '2-digit', minute: '2-digit' })}
            </Text>
          ))}
      </div>
    </div>
  )
}
