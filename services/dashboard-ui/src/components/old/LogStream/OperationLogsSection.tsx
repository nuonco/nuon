'use client'

import React, { type FC } from 'react'
import { LogsControls } from './LogsControls'
import { LogsModal } from './LogsModal'
import { LogsPreview } from './LogsPreview'
import { LogsViewerProvider } from './logs-viewer-context'
import { useLogs } from './logs-context'
import { Section } from '@/components/old/Card'
import { Loading } from '@/components/old/Loading'
import { Notice } from "@/components/old/Notice"

export interface IOperationLogsSection {
  actions?: React.ReactNode
  heading: React.ReactNode
}

export const OperationLogsSection: FC<IOperationLogsSection> = ({
  actions,
  heading,
}) => {
  const { error, isLoading, logs } = useLogs()
  
  return (
    <LogsViewerProvider>
      <Section
        heading={heading}
        actions={
          !error && logs?.length  ? (
            <div className="flex items-center divide-x">
              {actions ? <div className="mr-4">{actions}</div> : null}
              <div className="pl-4">
                <LogsControls />
              </div>
              <div className="ml-4 pl-4">
                <LogsModal heading={heading} logs={logs} />
              </div>
            </div>
          ) : null
        }
      >
        {error ? (
          <Notice variant={error?.message === "Log stream not created yet." ? "warn" : "error" }>
            {error?.message || "Unable to load log stream."}
          </Notice>
        ) : !logs?.length && isLoading ? (
          <div className="mt-12">
            <Loading loadingText="Loading logs..." variant="stack" />
          </div>
        ) : (
          <LogsPreview logs={logs} />
        )}
      </Section>
    </LogsViewerProvider>
  )
}
