'use client'

import { useEffect, useRef, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useFullUrl } from '@/hooks/use-full-url'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { getSeverityBorderClasses } from '@/utils/log-stream-utils'
import { AttributesTabs } from './AttributesTabs'
import { LogPanelHeading } from './LogPanelHeading'
import { LogMetadata } from './LogMetadata'

interface ILogPanel extends Omit<IPanel, 'heading' | 'children' | 'size'> {
  log: TOTELLog
  cycleDirection?: 'up' | 'down'
}

const LogPanelBody = ({ log, url }: { log: TOTELLog; url: string }) => (
  <>
    <div className="flex flex-col gap-2">
      <span className="flex items-center gap-2 justify-between">
        <Text weight="strong">Log details</Text>

        <Text variant="subtext" flex className="gap-2">
          Copy link
          <ClickToCopyButton textToCopy={url} />
        </Text>
      </span>
      <div className="flex flex-wrap gap-8">
        <LabeledValue label="Service">
          <Text>{log.service_name}</Text>
        </LabeledValue>

        <LabeledValue label="Scope">
          <Text>{log.scope_name}</Text>
        </LabeledValue>
      </div>
    </div>
    <div className="flex flex-col gap-2">
      <span className="flex items-center gap-2 justify-between">
        <Text weight="strong">Log message</Text>

        <ClickToCopyButton textToCopy={log.body} />
      </span>

      <Code className="!shadow-none">{log.body}</Code>
    </div>

    <AttributesTabs log={log} />

    <LogMetadata log={log} />
  </>
)

const LogPanelNavHint = () => (
  <div className="flex w-full mt-auto">
    <Text flex nowrap className="gap-2">
      Use
      <span className="inline-flex items-center gap-1">
        <Badge variant="code" size="sm">
          <Icon variant="ArrowUp" />
        </Badge>
        /
        <Badge variant="code" size="sm">
          <Icon variant="ArrowDown" />
        </Badge>
        or
        <Badge variant="code" size="sm">
          <Text family="mono" variant="subtext">
            k
          </Text>
        </Badge>
        /
        <Badge variant="code" size="sm">
          <Text family="mono" variant="subtext">
            j
          </Text>
        </Badge>
      </span>
      to navigate between logs.
    </Text>
  </div>
)

type TCycleTransition = {
  prevLog: TOTELLog
  direction: 'up' | 'down'
}

export const LogPanel = ({ className, log, cycleDirection, ...props }: ILogPanel) => {
  const url = useFullUrl()
  const prevLogRef = useRef<TOTELLog>(log)
  const [transition, setTransition] = useState<TCycleTransition | null>(null)

  useEffect(() => {
    if (log.id !== prevLogRef.current.id && cycleDirection) {
      setTransition({ prevLog: prevLogRef.current, direction: cycleDirection })
      prevLogRef.current = log
      const timer = setTimeout(() => setTransition(null), 150)
      return () => clearTimeout(timer)
    }
    prevLogRef.current = log
  }, [log.id, cycleDirection])

  const enterClass =
    transition?.direction === 'down'
      ? 'animate-slide-in-from-bottom'
      : transition?.direction === 'up'
        ? 'animate-slide-in-from-top'
        : undefined

  const exitClass =
    transition?.direction === 'down'
      ? 'animate-slide-exit-up'
      : transition?.direction === 'up'
        ? 'animate-slide-exit-down'
        : undefined

  return (
    <Panel
      className={cn(
        'border-t-6',
        getSeverityBorderClasses(log.severity_number, 't'),
        className
      )}
      heading={
        <div className="relative overflow-hidden">
          {transition && (
            <div className={cn('absolute inset-0', exitClass)}>
              <LogPanelHeading log={transition.prevLog} />
            </div>
          )}
          <div className={enterClass}>
            <LogPanelHeading log={log} />
          </div>
        </div>
      }
      size="half"
      {...props}
    >
      <div className="relative overflow-hidden flex flex-col flex-auto">
        {transition && (
          <div
            className={cn(
              'absolute inset-x-0 top-0 flex flex-col gap-4 md:gap-6',
              exitClass
            )}
          >
            <LogPanelBody log={transition.prevLog} url={url} />
          </div>
        )}
        <div className={cn('flex flex-col flex-auto gap-4 md:gap-6', enterClass)}>
          <LogPanelBody log={log} url={url} />
        </div>
      </div>
      <LogPanelNavHint />
    </Panel>
  )
}
