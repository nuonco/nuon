import { Text, type IText } from '@/components/common/Text'
import type { TAccount, TInstallActionRun } from '@/types'

export const actorName = (actor?: TAccount) =>
  actor?.name || actor?.email || actor?.id

export const getTriggerLabel = (trigger?: TInstallActionRun['trigger']) => {
  if (!trigger) return undefined

  const actor = actorName(trigger.actor)

  switch (trigger.type) {
    case 'cron':
      return 'system (cron)'
    case 'system':
      return 'system'
    case 'runbook': {
      const runbook = trigger.parent_name
        ? `runbook ${trigger.parent_name}`
        : 'runbook'
      return actor ? `${runbook} · started by ${actor}` : runbook
    }
    default:
      return actor ? `${trigger.type} · ${actor}` : trigger.type
  }
}

export const getTriggeredByLabel = (run?: TInstallActionRun) =>
  getTriggerLabel(run?.trigger)

export interface ITriggeredBy extends Pick<IText, 'variant' | 'theme'> {
  run?: TInstallActionRun
}

export const TriggeredBy = ({ run, ...textProps }: ITriggeredBy) => {
  const label = getTriggeredByLabel(run)

  return <Text {...textProps}>{label ?? '—'}</Text>
}
