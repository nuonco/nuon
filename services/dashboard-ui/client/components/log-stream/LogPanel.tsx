'use client'

import { useLayoutEffect, useRef, useState } from 'react'
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
          <Icon variant="ArrowUpIcon" />
        </Badge>
        /
        <Badge variant="code" size="sm">
          <Icon variant="ArrowDownIcon" />
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

const ANIM_DURATION = 100

type TCycleTransition = {
  phase: 'exiting' | 'entering'
  prevLog: TOTELLog
  direction: 'up' | 'down'
}

const LOG_PANEL_FULLSCREEN_KEY = 'log-panel-fullscreen'

export const LogPanel = ({ className, log, cycleDirection, ...props }: ILogPanel) => {
  const url = useFullUrl()
  const prevLogRef = useRef<TOTELLog>(log)
  const isFullscreen = localStorage.getItem(LOG_PANEL_FULLSCREEN_KEY) === 'true'
  const [transition, setTransition] = useState<TCycleTransition | null>(null)

  useLayoutEffect(() => {
    if (log.id !== prevLogRef.current.id && cycleDirection) {
      const prevLog = prevLogRef.current
      prevLogRef.current = log
      setTransition({ phase: 'exiting', prevLog, direction: cycleDirection })
      const exitTimer = setTimeout(() => {
        requestAnimationFrame(() => {
          setTransition({ phase: 'entering', prevLog, direction: cycleDirection })
        })
      }, ANIM_DURATION)
      const enterTimer = setTimeout(() => {
        requestAnimationFrame(() => {
          setTransition(null)
        })
      }, ANIM_DURATION * 2)
      return () => {
        clearTimeout(exitTimer)
        clearTimeout(enterTimer)
      }
    }
    prevLogRef.current = log
  }, [log.id, cycleDirection])

  const isExiting = transition?.phase === 'exiting'
  const isEntering = transition?.phase === 'entering'

  const exitClass =
    transition?.direction === 'down'
      ? 'animate-slide-exit-up'
      : transition?.direction === 'up'
        ? 'animate-slide-exit-down'
        : undefined

  const enterClass =
    transition?.direction === 'down'
      ? 'animate-slide-in-from-bottom'
      : transition?.direction === 'up'
        ? 'animate-slide-in-from-top'
        : undefined

  const displayLog = isExiting ? transition.prevLog : log

  return (
    <Panel
      className={cn(
        'border-t-6',
        getSeverityBorderClasses(
          (isExiting ? transition.prevLog : log).severity_number,
          't'
        ),
        className
      )}
      heading={
        <div className={cn(
          isExiting && 'animate-heading-exit',
          isEntering && 'animate-heading-enter'
        )}>
          <LogPanelHeading log={displayLog} />
        </div>
      }
      size="half"
      defaultExpanded={isFullscreen}
      onSizeChange={(size) => {
        localStorage.setItem(LOG_PANEL_FULLSCREEN_KEY, size === 'full' ? 'true' : 'false')
      }}
      {...props}
    >
      <div className="relative flex flex-col flex-auto">
        <div className={cn(
          'flex flex-col flex-auto gap-4 md:gap-6',
          isExiting && exitClass,
          isEntering && enterClass
        )}>
          <LogPanelBody log={displayLog} url={url} />
        </div>
      </div>
      <LogPanelNavHint />
    </Panel>
  )
}
