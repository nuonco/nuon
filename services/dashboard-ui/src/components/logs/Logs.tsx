import classNames from 'classnames'
import React, { type FC } from 'react'
import { Code } from '@/components'

export type TWaypointLogTerminalEvent = {
  line?: { msg: string }
  raw?: { data: string }
  step?: { msg: string }
  status?: { msg: string }
}

export type TWaypointLog = {
  Complete: Record<string, unkown>
  Open: Record<string, unknown>
  State: Record<string, unknown>
  Terminal: {
    buffered: boolean
    events: Array<Record<string, unknown>>
  }
}

export interface ILogs {
  logs: Array<TWaypointLog>
}

export const Logs: FC<ILogs> = ({ logs }) => { 
  return (
    <Code>
      {logs?.length
        ? logs.map((term) => {
            // handle complete state

            return (<>{term?.Terminal?.events?.length
              ? term?.Terminal?.events?.map((l, i) => {
                  let line = null

                  if (l?.line) {
                    line = (
                      <span
                        key={`${l?.line?.msg}-${i}`}
                        className="block text-xs"
                      >
                        {l?.line?.msg}
                      </span>
                    )
                  }

                  // raw data

                  if (l?.raw?.data) {
                    line = (
                      <span
                        key={`${l?.raw?.data}-${i}`}
                        className="block text-xs"
                      >
                        {atob(l?.raw?.data)}
                      </span>
                    )
                  }

                  if (l?.step) {
                    line = (
                      <span
                        key={`${l?.step?.msg}-${i}`}
                        className="block text-xs"
                      >
                        {l?.step?.msg}
                      </span>
                    )
                  }

                  // status
                  if (l?.status) {
                    line = (
                      <span
                        key={`${l?.status?.msg}-${i}`}
                        className="block text-xs"
                      >
                        {l?.status?.msg}
                      </span>
                    )
                  }

                  return line
                })
              : null} {term?.State?.current}</>)
          })
        : 'no logs to show'}
    </Code>
  )
}
