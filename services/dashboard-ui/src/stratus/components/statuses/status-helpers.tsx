import type { TIconVariant } from "@/stratus/components/common"

type TStatusTheme = 'success' | 'warn' | 'neutral' | 'error' | 'info' | 'brand'

export function getStatusTheme(status: string): TStatusTheme {
  let theme: TStatusTheme

  switch (status) {
    case 'active':
    case 'ok':
    case 'finished':
    case 'healthy':
    case 'connected':
    case 'approved':
      theme = 'success'
      break
    case 'failed':
    case 'error':
    case 'bad':
    case 'access-error':
    case 'access_error':
    case 'timed-out':
    case 'unknown':
    case 'unhealthy':
    case 'not connected':
    case 'not-connected':
    case 'timed-out':
      theme = 'error'
      break
    case 'approval-denied':
    case 'approval-waiting':
    case 'cancelled':
    case 'outdated':
      theme = 'warn'
      break
    case 'executing':
    case 'waiting':
    case 'started':
    case 'in-progress':
    case 'building':
    case 'queued':
    case 'planning':
    case 'provisioning':
    case 'syncing':
    case 'deploying':
    case 'available':
    case 'pending-approval':
      theme = 'info'
      break
    case 'noop':
    case 'inactive':
    case 'pending':
    case 'offline':
    case 'Not deployed':
    case 'No build':
    case 'not-attempted':
    case 'deprovisioned':
      theme = 'neutral'
      break
    case 'special':
      theme = 'brand'
      break
    default:
      theme = 'neutral'
  }
  return theme
}

export function getStatusIconVariant(status: string): TIconVariant {
  let icon: TIconVariant

  switch (status) {
    case 'active':
    case 'ok':
    case 'finished':
    case 'healthy':
    case 'connected':
    case 'approved':
      icon = "CheckCircle"
      break
    case 'failed':
    case 'error':
    case 'bad':
    case 'access-error':
    case 'access_error':
    case 'timed-out':
    case 'unknown':
    case 'unhealthy':
    case 'not connected':
    case 'timed-out':
      icon = "XCircle"
      break
    case 'approval-denied':
    case 'approval-waiting':
    case 'cancelled':
    case 'outdated':
      icon = "Warning"
      break
    case 'executing':
    case 'waiting':
    case 'started':
    case 'in-progress':
    case 'building':
    case 'queued':
    case 'planning':
    case 'provisioning':
    case 'syncing':
    case 'deploying':
    case 'available':
    case 'pending-approval':
      icon = "WarningDiamond"
      break
    case 'noop':
    case 'inactive':
    case 'pending':
    case 'offline':
    case 'Not deployed':
    case 'No build':
    case 'not-attempted':
    case 'deprovisioned':
      icon = "ClockCountdown"
      break
    case 'special':
      icon = "Prohibit"
      break
    default:
      icon = "ClockCountdown"
  }

  return icon
}
