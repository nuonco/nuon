import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Text } from '../atoms/Text'
import { BranchProfile } from './BranchProfile'

export default {
  title: 'lite/molecules/BranchProfile',
}

const BRANCH = {
  id: 'brncq7fplr1up5atx5zpxotbabm',
  name: 'main',
  configs: [
    {
      config_number: 14,
      connected_github_vcs_config: {
        repo: 'acme/payments',
        branch: 'main',
        directory: 'services/payments',
      },
    },
  ],
}

export const Overview = () => (
  <ComponentDocs
    name="BranchProfile"
    tier="molecule"
    summary="A branch's identity: git glyph, branch name with its config directory, and the branch ID."
    use={[
      'Identify a branch in a switcher row, a menu, or a header.',
      'Render it with loading set while the branch resolves.',
    ]}
    avoid={[
      'Do not use it for branch status, run history, or install counts.',
      'Do not wrap the git glyph in an avatar; a branch has no image.',
    ]}
    rules={[
      'The directory comes from the latest branch config and is omitted when unset.',
      'The name renders in mono so it reads as a git ref.',
      'The ID is not copyable here; rows are navigation targets.',
      'Loading keeps the glyph and every line, so rows do not shift when data lands.',
      'The modeline variant is a single row for the status bar: glyph and name only.',
    ]}
    props={[
      {
        name: 'branch',
        type: 'TAppBranch',
        description: 'Branch to identify. Falls back to an unavailable label.',
      },
      {
        name: 'variant',
        type: "'full' | 'modeline'",
        default: 'full',
        description:
          'full stacks the name over the ID; modeline is one compact status bar row.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows name, directory, and ID loading shapes.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <BranchProfile branch={BRANCH} />
  </div>
)

export const WithoutDirectory = () => (
  <div className="p-8">
    <BranchProfile branch={{ id: 'brncq7fplr1up5atx5zpxotbabm', name: 'main' }} />
  </div>
)

export const LongValues = () => (
  <div className="w-64 p-8">
    <BranchProfile
      branch={{
        id: 'brncq7fplr1up5atx5zpxotbabm',
        name: 'feature/long-running-migration-branch',
        configs: [
          {
            config_number: 3,
            connected_github_vcs_config: {
              directory: 'services/payments/infrastructure/terraform',
            },
          },
        ],
      }}
    />
  </div>
)

export const Missing = () => (
  <div className="p-8">
    <BranchProfile />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <BranchProfile loading />
  </div>
)

export const Modeline = () => (
  <div className="flex h-7 w-fit items-center gap-2 bg-surface-02 px-3">
    <Text variant="caption" family="mono" weight="medium">
      payments
    </Text>
    <Text variant="caption" color="tertiary" aria-hidden>
      /
    </Text>
    <BranchProfile branch={BRANCH} variant="modeline" />
  </div>
)

export const ModelineLoading = () => (
  <div className="flex h-7 w-fit items-center gap-2 bg-surface-02 px-3">
    <Text variant="caption" family="mono" loading loadingWidth={12} />
    <Text variant="caption" color="tertiary" aria-hidden>
      /
    </Text>
    <BranchProfile loading variant="modeline" />
  </div>
)

export const Rows = () => (
  <div className="flex w-72 flex-col gap-0.5 p-8">
    {['main', 'release', 'preview'].map((name, index) => (
      <BranchProfile
        key={name}
        branch={{
          id: `brncq7fplr1up5atx5zpxotbab${index}`,
          name,
          configs: [
            {
              config_number: 14 - index,
              connected_github_vcs_config: { directory: 'services/payments' },
            },
          ],
        }}
      />
    ))}
  </div>
)
