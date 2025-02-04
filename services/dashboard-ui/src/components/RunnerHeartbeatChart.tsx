'use client'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { DateTime } from 'luxon'
import { type FC } from 'react'
import { Line } from 'react-chartjs-2'
import type { TRunnerRecentHeartbeat } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

export const options = {}

export const RunnerHeartbeatChart: FC<{
  heartBeats: Array<TRunnerRecentHeartbeat>
}> = ({ heartBeats }) => {
  const parsedData = heartBeats?.reduce(
    (acc, item) => {
      acc['labels'].push(
        DateTime.fromISO(item?.truncated_time).toLocaleString(
          DateTime.TIME_SIMPLE
        )
      )
      acc['data'].push(item?.record_count)

      return acc
    },
    { labels: [], data: [] }
  )

  const data = {
    labels: parsedData['labels'],
    datasets: [
      {
        label: 'Heartbeat count',
        data: parsedData['data'],
        borderColor: '#8040BF',
        backgroundColor: '#F2E5FF',
        stepped: true,
      },
    ],
  }

  return (
    <Line
      options={{
        responsive: true,
        plugins: {
          legend: {
            position: 'top' as const,
            display: false,
          },
          title: {
            display: false,
            text: 'Heartbeat records',
          },
        },

        scales: {
          x: {
            ticks: {
              //              maxTicksLimit: 30,
            },
            grid: {
              color: '#CFD6DD66',
            },
          },
          y: {
            ticks: {
              // maxTicksLimit: 2,
            },
            grid: {
              color: '#CFD6DD66',
            },
            max: 12,
            min: 0,
          },
        },
      }}
      data={data}
    />
  )
}
