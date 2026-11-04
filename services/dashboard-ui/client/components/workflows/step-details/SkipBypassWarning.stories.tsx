export default {
  title: 'Features / Workflows / Step details / Skip bypass warning',
}

import { SkipBypassWarning } from './SkipBypassWarning'
import type { TWorkflowStep } from '@/types'

const ranAnyway = {
  id: 'apply',
  name: 'teardown apply plan database',
  idx: 101,
} as TWorkflowStep

export const Default = () => <SkipBypassWarning ranAnyway={ranAnyway} />
