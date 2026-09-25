export default { title: 'Cloud connections/Concepts/Runbook checklist' }

import { RunbookChecklist } from './RunbookChecklist'

export const Fresh = () => <RunbookChecklist />
export const Midway = () => <RunbookChecklist completedThrough={2} />
export const VerificationFailedAtPermissions = () => (
  <RunbookChecklist completedThrough={2} failedStep={3} />
)
export const AllDone = () => <RunbookChecklist completedThrough={5} allDone />

Fresh.meta = { fullBleed: true }
Midway.meta = { fullBleed: true }
VerificationFailedAtPermissions.meta = { fullBleed: true }
AllDone.meta = { fullBleed: true }
