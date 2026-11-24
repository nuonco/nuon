import { Text } from '@/components/common/Text'
import './WaitingForLogs.css'

export const WaitingForLogs = () => (
  <div
    role="status"
    aria-live="polite"
    className="mt-3 rounded-md border bg-code p-4"
  >
    <Text family="mono" variant="subtext">
      Waiting on logs
      <span aria-hidden>
        <span className="waiting-dot">.</span>
        <span className="waiting-dot">.</span>
        <span className="waiting-dot">.</span>
      </span>
    </Text>
  </div>
)
