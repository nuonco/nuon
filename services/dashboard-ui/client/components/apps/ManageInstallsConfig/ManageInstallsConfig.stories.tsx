import { ModalStory } from '@/components/__stories__/helpers'
import { ManageInstallsConfigComponent } from './index'

export default {
  title: 'Apps/ManageInstallsConfig',
}

const mockRepos = [
  {
    id: 1,
    name: 'app-configs',
    full_name: 'acme/app-configs',
    private: true,
    fork: false,
    html_url: 'https://github.com/acme/app-configs',
    default_branch: 'main',
    updated_at: '2026-01-01',
  },
  {
    id: 2,
    name: 'infra',
    full_name: 'acme/infra',
    private: true,
    fork: false,
    html_url: 'https://github.com/acme/infra',
    default_branch: 'main',
    updated_at: '2026-01-01',
  },
]

const mockBranches = [
  { name: 'main' },
  { name: 'staging' },
  { name: 'develop' },
]

const noop = () => {}

export const NewConfig = () => (
  <ModalStory>
    <ManageInstallsConfigComponent
      vcsConnections={[]}
      repos={mockRepos}
      branches={mockBranches}
      loadingRepos={false}
      loadingBranches={false}
      reposError={null}
      branchesError={null}
      selectedVcsConnectionId=""
      onVcsConnectionChange={noop}
      selectedRepo={null}
      onRepoChange={noop}
      selectedBranch=""
      onBranchChange={noop}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const WithExistingConfig = () => (
  <ModalStory>
    <ManageInstallsConfigComponent
      currentConfig={{
        id: 'aic123',
        created_at: '2026-07-01T00:00:00Z',
        app_id: 'app123',
        vcs_type: 'connected',
        repo: 'acme/app-configs',
        branch: 'main',
        directory: 'install-configs',
        source: 'config',
      }}
      vcsConnections={[]}
      repos={mockRepos}
      branches={mockBranches}
      loadingRepos={false}
      loadingBranches={false}
      reposError={null}
      branchesError={null}
      selectedVcsConnectionId=""
      onVcsConnectionChange={noop}
      selectedRepo={mockRepos[0]}
      onRepoChange={noop}
      selectedBranch="main"
      onBranchChange={noop}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)
