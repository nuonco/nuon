'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
import type { TSandboxRun } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface ISandboxHistory {
  installId: string
  initSandboxRuns: Array<TSandboxRun>
  orgId: string
  shouldPoll?: boolean
}

export const SandboxHistory: FC<ISandboxHistory> = ({
  installId,
  initSandboxRuns,
  shouldPoll = false,
  orgId,
}) => {
  const [sandboxRuns, setSandboxRuns] = useState(initSandboxRuns)

  useEffect(() => {
    const fetchSandboxRuns = () => {
      fetch(`/api/${orgId}/installs/${installId}/sandbox-runs`)
        .then((res) => res.json().then((r) => setSandboxRuns(r)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollSandboxRuns = setInterval(fetchSandboxRuns, SHORT_POLL_DURATION)
      return () => clearInterval(pollSandboxRuns)
    }
  }, [sandboxRuns, orgId, shouldPoll])

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No runs yet"
          emptyMessage="Waiting on sandbox runs."
          variant="history"
          isSmall
        />
      }
      events={sandboxRuns.map((run, i) => ({
        id: run.id,
        status: run.status,
        underline: (
          <div>
            <Text>
              <Text variant="mono-12">
                <ToolTip tipContent={run.id}>
                  <Truncate variant="small">{run.id}</Truncate>
                </ToolTip>
              </Text>
              <>
                /{' '}
                {run?.run_type.length >= 12 ? (
                  <ToolTip tipContent={run?.run_type} alignment="right">
                    <Truncate variant="small">{run?.run_type}</Truncate>
                  </ToolTip>
                ) : (
                  run.run_type
                )}
              </>
            </Text>
            {run?.created_by ? (
              <Text className="text-cool-grey-600 dark:text-white/70 !text-[10px]">
                Run by: {run?.created_by?.email}
              </Text>
            ) : null}
          </div>
        ),
        time: run.updated_at,
        href: `/${orgId}/installs/${installId}/sandbox/${run.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}
