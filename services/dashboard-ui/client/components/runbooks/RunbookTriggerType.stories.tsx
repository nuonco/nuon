export default {
  title: 'Features / Runbooks / Trigger type',
}

import { RunbookTriggerType } from './RunbookTriggerType'

export const Manual = () => <RunbookTriggerType triggerType="manual" />

export const Cron = () => (
  <RunbookTriggerType triggerType="cron" cronSchedule="0 9 * * 1-5" />
)
