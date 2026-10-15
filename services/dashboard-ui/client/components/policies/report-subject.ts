import type { TPolicyReport } from '@/types'

const SUBJECT_PREFIX: Record<string, string> = {
  install_deploys: 'Deploy',
  component_builds: 'Build',
}

export const reportSubject = (report: TPolicyReport): string => {
  if (!report?.component_name) return 'Sandbox'
  const prefix = SUBJECT_PREFIX[report.owner_type ?? '']
  return prefix ? `${prefix}: ${report.component_name}` : report.component_name
}
