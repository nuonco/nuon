import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { AppBranchOverviewCards } from './AppBranchOverviewCards'

export default {
  title: 'lite/organisms/AppBranchOverviewCards',
}

export const Overview = () => (
  <ComponentDocs
    name="AppBranchOverviewCards"
    tier="organism"
    summary="The three-card summary at the top of an app branch overview: branch info, last update, and installs."
    use={[
      'Place at the top of the app branch overview, above the rest of the page.',
      'Use the loading state while the branch and its install count resolve.',
    ]}
    avoid={[
      'Do not add a fourth card; the app branch summary is three cards.',
      'Do not use it for install-scoped facts such as health or drift.',
    ]}
    rules={[
      'Branch info reads from the latest branch config, falling back to the VCS directory when no repository is set.',
      'Last update reads the latest run commit and renders nothing but an empty state when the branch has never run.',
      'An install count that is capped by pagination renders with a trailing plus.',
    ]}
    props={[
      {
        name: 'branch',
        type: 'TAppBranch',
        description: 'Branch whose config and latest run are summarized.',
      },
      {
        name: 'installCount',
        type: 'number',
        description:
          'Installs assigned to the branch. Loads until it is defined.',
      },
      {
        name: 'hasMoreInstalls',
        type: 'boolean',
        default: 'false',
        description:
          'Marks the count as a lower bound when more installs remain unfetched.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows card loading shapes while keeping every title.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <AppBranchOverviewCards
      branch={{
        id: 'br_main',
        name: 'main',
        configs: [
          {
            config_number: 14,
            connected_github_vcs_config: {
              repo: 'acme/payments',
              branch: 'main',
            },
          },
        ],
        latest_run: {
          updated_at: '2026-09-08T13:45:00Z',
          vcs_connection_commit: {
            sha: 'a1b2c3d4e5f6',
            message: 'Update component versions',
            author_name: 'Alex Smith',
            created_at: '2026-09-08T13:42:00Z',
          },
        },
      }}
      installCount={12}
    />
  </div>
)

export const Empty = () => (
  <div className="p-8">
    <AppBranchOverviewCards branch={{ id: 'br_main', name: 'main' }} installCount={0} />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <AppBranchOverviewCards loading />
  </div>
)
