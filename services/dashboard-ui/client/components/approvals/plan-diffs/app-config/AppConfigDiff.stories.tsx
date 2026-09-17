export default {
  title: 'Approvals/PlanDiffs/AppConfigDiff',
}

import {
  appConfigAllSections,
  appConfigMockSections,
} from '@/lib/fixtures/plan-diffs/app-config'
import { AppConfigDiff } from './AppConfigDiff'

export const Default = () => (
  <AppConfigDiff
    sections={appConfigMockSections}
    summary={{ added: 2, removed: 1, changed: 3 }}
  />
)

export const NoChanges = () => <AppConfigDiff sections={[]} summary={null} />

export const Loading = () => (
  <AppConfigDiff sections={[]} summary={null} isLoading />
)

export const ComponentsOnly = () => (
  <AppConfigDiff
    sections={[appConfigMockSections[0]]}
    summary={{ added: 1, removed: 1, changed: 1 }}
  />
)

export const FieldsOnly = () => (
  <AppConfigDiff
    sections={[appConfigMockSections[2]]}
    summary={{ added: 0, removed: 0, changed: 2 }}
  />
)

export const AllSections = () => (
  <AppConfigDiff
    sections={appConfigAllSections}
    summary={{ added: 6, removed: 1, changed: 5 }}
  />
)
