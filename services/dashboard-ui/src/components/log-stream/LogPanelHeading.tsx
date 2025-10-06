import { Badge } from '@/components/common/Badge'
import { Time } from '@/components/common/Time'
import type { TOTELLog } from '@/types'
import { getBadgeThemeFromSeverity } from '@/utils/log-stream-utils'

interface ILogPanelHeading {
  log: TOTELLog
}

export const LogPanelHeading = ({ log }: ILogPanelHeading) => {
  return (
    <div className="flex items-center gap-4">
      <Badge
        variant="code"
        theme={getBadgeThemeFromSeverity(log.severity_number)}
      >
        {log.severity_text}
      </Badge>

      <Time
        time={log.timestamp}
        format="log-datetime"
        variant="base"
        weight="strong"
      />

      <Time
        time={log.timestamp}
        format="relative"
        variant="subtext"
        theme="neutral"
      />
    </div>
  )
}
