import {
  appConfigAllSections,
  appConfigMockSections,
} from '@/lib/fixtures/plan-diffs/app-config'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { AppConfigDiff } from './AppConfigDiff'

export default {
  title: 'lite/organisms/AppConfigDiff',
}

const Frame = (props: Parameters<typeof AppConfigDiff>[0]) => (
  <div className="max-w-6xl p-8">
    <AppConfigDiff {...props} />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="AppConfigDiff"
    tier="organism"
    summary="App configuration changes normalized into searchable TOML and embedded file diffs."
    use={[
      'Review component, action, runbook, input, secret, runner, permission, stack, and sandbox changes.',
      'Keep embedded files searchable alongside their owning entity.',
    ]}
    avoid={[
      'Do not flatten embedded file contents into field labels.',
      'Do not rebuild legacy fixtures for lite stories.',
    ]}
    rules={[
      'Section content takes precedence over derived field rows.',
      'Embedded files retain their own language and filename.',
      'Every legacy app-config story uses the exact same shared fixture.',
      'Pass defaultSectionsOpen to start every section expanded; omit it to use the reader\'s plan preference.',
    ]}
    props={[
      {
        name: 'sections',
        type: 'TAppConfigDiffSection[]',
        description: 'Normalized app-config sections.',
      },
      {
        name: 'summary',
        type: 'IAppConfigDiffSummary | null',
        description: 'Legacy add, change, and remove counts.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Displays the loading state.',
      },
      {
        name: 'defaultSectionsOpen',
        type: 'boolean',
        description:
          'Initial expand state for every section. Omitted, the plan preference decides.',
      },
    ]}
  />
)

export const Default = () => (
  <Frame
    sections={appConfigMockSections}
    summary={{ added: 2, removed: 1, changed: 3 }}
  />
)

export const NoChanges = () => <Frame sections={[]} summary={null} />

export const Loading = () => <Frame sections={[]} summary={null} loading />

export const ComponentsOnly = () => (
  <Frame
    sections={[appConfigMockSections[0]]}
    summary={{ added: 1, removed: 1, changed: 1 }}
  />
)

export const FieldsOnly = () => (
  <Frame
    sections={[appConfigMockSections[2]]}
    summary={{ added: 0, removed: 0, changed: 2 }}
  />
)

export const AllSections = () => (
  <Frame
    sections={appConfigAllSections}
    summary={{ added: 6, removed: 1, changed: 5 }}
  />
)
