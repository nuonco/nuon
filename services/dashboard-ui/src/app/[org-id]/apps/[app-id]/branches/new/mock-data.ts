import { IInstall } from './install-card'
import { IGroupTemplate, IAppBranchConfig } from './types'

export interface IMockRepository {
  id: string
  full_name: string
  name: string
  owner: string
  private: boolean
  description?: string
}

export interface IMockBranch {
  name: string
  protected: boolean
  default: boolean
}

export const mockVCSConnections = [
  { id: 'vcs1', name: 'github-org', type: 'github' },
  { id: 'vcs2', name: 'github-personal', type: 'github' },
]

export const mockRepos = [
  { id: 'repo1', name: 'nuonco/app-backend', private: true },
  { id: 'repo2', name: 'nuonco/app-frontend', private: false },
  { id: 'repo3', name: 'nuonco/infrastructure', private: true },
]

export const mockBranches = ['main', 'develop', 'staging', 'production']

/**
 * Get mock repositories for testing VCS integration
 * Used as fallback when API calls fail
 */
export function getMockRepositories(): IMockRepository[] {
  return [
    {
      id: 'repo-001',
      full_name: 'nuonco/platform',
      name: 'platform',
      owner: 'nuonco',
      private: true,
      description: 'Main platform monorepo',
    },
    {
      id: 'repo-002',
      full_name: 'nuonco/dashboard-ui',
      name: 'dashboard-ui',
      owner: 'nuonco',
      private: true,
      description: 'Next.js dashboard application',
    },
    {
      id: 'repo-003',
      full_name: 'nuonco/infrastructure',
      name: 'infrastructure',
      owner: 'nuonco',
      private: true,
      description: 'Terraform and Kubernetes configs',
    },
    {
      id: 'repo-004',
      full_name: 'acme-corp/backend-api',
      name: 'backend-api',
      owner: 'acme-corp',
      private: false,
      description: 'Public REST API',
    },
    {
      id: 'repo-005',
      full_name: 'acme-corp/frontend',
      name: 'frontend',
      owner: 'acme-corp',
      private: false,
      description: 'React frontend application',
    },
  ]
}

/**
 * Get mock branches for a specific repository
 * Returns repo-specific branches based on repository name
 */
export function getMockBranches(repoFullName: string): IMockBranch[] {
  const commonBranches: IMockBranch[] = [
    { name: 'main', protected: true, default: true },
    { name: 'develop', protected: false, default: false },
    { name: 'staging', protected: false, default: false },
  ]

  // Add repo-specific branches
  if (repoFullName.includes('platform')) {
    return [
      ...commonBranches,
      { name: 'production', protected: true, default: false },
      { name: 'feature/new-deployment', protected: false, default: false },
    ]
  }

  if (repoFullName.includes('infrastructure')) {
    return [
      ...commonBranches,
      { name: 'terraform-updates', protected: false, default: false },
    ]
  }

  if (repoFullName.includes('backend-api')) {
    return [
      ...commonBranches,
      { name: 'release/v2.0', protected: true, default: false },
      { name: 'hotfix/critical-bug', protected: false, default: false },
    ]
  }

  return commonBranches
}

export const mockInstalls: IInstall[] = [
  {
    id: 'ins1',
    name: 'production-us-east',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins2',
    name: 'production-us-west',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins3',
    name: 'staging-us-east',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins4',
    name: 'staging-us-west',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins5',
    name: 'dev-environment',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins6',
    name: 'qa-environment',
    region: 'eu-west-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins7',
    name: 'demo-environment',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins8',
    name: 'sandbox-1',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins9',
    name: 'sandbox-2',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins10',
    name: 'test-env',
    region: 'us-east-1',
    status: 'inactive' as const,
    platform: 'aws' as const,
  },
]

export function getMockInstalls(): IInstall[] {
  return mockInstalls
}

export function getDefaultGroupTemplates(): IGroupTemplate[] {
  return [
    {
      id: 'prod-staging',
      name: 'Production → Staging',
      description: 'Deploy to staging first, then production after approval',
      groups: [
        {
          name: 'Staging',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 5,
        },
        {
          name: 'Production',
          requiresApproval: true,
          rollbackOnFailure: true,
          maxParallel: 2,
        },
      ],
    },
    {
      id: 'canary',
      name: 'Canary Deployment',
      description: 'Deploy to canary, then gradually roll out to all installs',
      groups: [
        {
          name: 'Canary',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 1,
        },
        {
          name: 'Wave 1',
          requiresApproval: true,
          rollbackOnFailure: true,
          maxParallel: 3,
        },
        {
          name: 'Wave 2',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 10,
        },
      ],
    },
    {
      id: 'regional',
      name: 'Regional Rollout',
      description: 'Deploy region by region for geographic rollout',
      groups: [
        {
          name: 'US East',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 5,
        },
        {
          name: 'US West',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 5,
        },
        {
          name: 'Europe',
          requiresApproval: false,
          rollbackOnFailure: true,
          maxParallel: 5,
        },
      ],
    },
    {
      id: 'custom',
      name: 'Custom',
      description: 'Start with empty groups and configure manually',
      groups: [],
    },
  ]
}

export function saveConfigToLocalStorage(
  branchId: string,
  config: IAppBranchConfig
): void {
  try {
    const key = `app-branch-config-${branchId}`
    localStorage.setItem(key, JSON.stringify(config))
  } catch (err) {
    console.error('Failed to save branch config to localStorage:', err)
  }
}

export function loadConfigFromLocalStorage(
  branchId: string
): IAppBranchConfig | null {
  try {
    const key = `app-branch-config-${branchId}`
    const data = localStorage.getItem(key)
    if (data) {
      return JSON.parse(data) as IAppBranchConfig
    }
  } catch (err) {
    console.error('Failed to load branch config from localStorage:', err)
  }
  return null
}

export function getAllBranchConfigs(): Record<string, IAppBranchConfig> {
  const configs: Record<string, IAppBranchConfig> = {}
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith('app-branch-config-')) {
        const branchId = key.replace('app-branch-config-', '')
        const data = localStorage.getItem(key)
        if (data) {
          configs[branchId] = JSON.parse(data)
        }
      }
    }
  } catch (err) {
    console.error('Failed to load all branch configs:', err)
  }
  return configs
}