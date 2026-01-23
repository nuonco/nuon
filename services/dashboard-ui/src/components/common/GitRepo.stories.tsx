import type { TVCSGit, TVCSGitHub } from '@/types/ctl-api.types'
import { GitRepo } from './GitRepo'

// Mock data for TVCSGitHub
const mockGitHubConfig: TVCSGitHub = {
  id: 'vcs-github-1',
  repo: 'https://github.com/acme-corp/api-service',
  repo_name: 'api-service',
  repo_owner: 'acme-corp',
  branch: 'main',
  directory: 'src/deployments',
  vcs_connection_id: 'conn-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:30:00Z',
}

// Mock data for TVCSGit
const mockGitConfig: TVCSGit = {
  id: 'vcs-git-1',
  repo: 'https://gitlab.com/myorg/backend-service.git',
  branch: 'develop',
  directory: 'infra/kubernetes',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:30:00Z',
}

// Mock data for minimal config
const mockMinimalConfig: TVCSGit = {
  id: 'vcs-git-2',
  repo: 'https://github.com/company/frontend',
  branch: 'main',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:30:00Z',
}

// Mock data with all GitHub fields
const mockAllFieldsConfig: TVCSGitHub = {
  id: 'vcs-github-2',
  repo: 'https://github.com/platform-team/infrastructure',
  repo_name: 'infrastructure',
  repo_owner: 'platform-team',
  branch: 'production',
  directory: 'terraform/environments/prod',
  vcs_connection_id: 'conn-456',
  component_config_id: 'comp-789',
  component_config_type: 'helm',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-20T14:45:00Z',
  created_by_id: 'user-123',
}

export const GitHubRepository = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">GitHub Repository</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Displays connected GitHub repository information with{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          repo_owner
        </code>
        ,{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          repo_name
        </code>
        ,{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          branch
        </code>
        , and{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          directory
        </code>{' '}
        fields.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Connected GitHub VCS Config</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockGitHubConfig} />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Repository:</strong> Full repository URL or path
      </div>
      <div>
        <strong>Owner:</strong> GitHub account or organization name
      </div>
      <div>
        <strong>Repository Name:</strong> Parsed repository name
      </div>
      <div>
        <strong>Branch:</strong> Git branch being tracked
      </div>
      <div>
        <strong>Directory:</strong> Subdirectory within the repository
      </div>
    </div>
  </div>
)

export const PublicGitRepository = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Public Git Repository</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Displays public git repository information from any git provider. Shows{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          repo
        </code>
        ,{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          branch
        </code>
        , and{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          directory
        </code>{' '}
        without GitHub-specific metadata.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Public Git VCS Config</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockGitConfig} />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Repository:</strong> Full git repository URL
      </div>
      <div>
        <strong>Branch:</strong> Git branch being tracked
      </div>
      <div>
        <strong>Directory:</strong> Subdirectory within the repository
      </div>
    </div>
  </div>
)

export const WithDirectory = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Repository With Subdirectory</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        When the{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          directory
        </code>{' '}
        field is populated, it displays the specific subdirectory within the
        repository that contains the component configuration.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">GitHub with Directory</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockGitHubConfig} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Public Git with Directory</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockGitConfig} />
      </div>
    </div>
  </div>
)

export const MinimalGitRepo = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Minimal Configuration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Shows the minimum required fields. Only{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          repo
        </code>{' '}
        and{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          branch
        </code>{' '}
        are displayed. Optional fields like{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          directory
        </code>{' '}
        are hidden when not provided.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Minimal Fields</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockMinimalConfig} />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Auto-hiding:</strong> Undefined fields are automatically hidden
      </div>
      <div>
        <strong>Responsive:</strong> Grid layout adapts to screen size
      </div>
    </div>
  </div>
)

export const AllFields = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">All Fields Populated</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Example showing all possible fields for a{' '}
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">
          TVCSGitHub
        </code>{' '}
        configuration. This represents the maximum information available for a
        connected GitHub repository.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complete GitHub VCS Config</h4>
      <div className="p-6 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <GitRepo vcsConfig={mockAllFieldsConfig} />
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>Type Detection:</strong> Automatically detects TVCSGitHub vs
        TVCSGit
      </div>
      <div>
        <strong>GitHub Fields:</strong> Shows owner and repo name separately
      </div>
      <div>
        <strong>Grid Layout:</strong> 2-column responsive grid on desktop
      </div>
      <div>
        <strong>LabeledValue:</strong> Uses LabeledValue component for
        consistency
      </div>
    </div>
  </div>
)
