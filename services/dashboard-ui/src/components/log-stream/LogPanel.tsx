'use client'

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
}

export const LogPanel = ({ className, log, ...props }: ILogPanel) => {
  const url = useFullUrl()

  return (
    <Panel
      className={cn(
        'border-t-6',
        getSeverityBorderClasses(log.severity_number, 't'),
        className
      )}
      panelKey={log.id}
      heading={<LogPanelHeading log={log} />}
      size="half"
      {...props}
    >
      <div className="flex flex-col gap-2">
        <span className="flex items-center gap-2 justify-between">
          <Text weight="strong">Log details</Text>

          <Text variant="subtext" className="flex items-center gap-2">
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

      <div className="flex w-full mt-auto">
        <Text className="inline-flex gap-2 items-center !text-nowrap">
          Use
          <span className="inline-flex items-center gap-1">
            <Badge variant="code" size="sm">
              <Icon variant="ArrowUp" />
            </Badge>
            /
            <Badge variant="code" size="sm">
              <Icon variant="ArrowDown" />
            </Badge>
          </span>
          to navigate between logs.
        </Text>
      </div>
    </Panel>
  )
}
