import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Modal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { useToast } from '@/hooks/use-toast'
import { createBranchConfig, updateBranch } from '@/lib'
import type { TAppBranch, TAppBranchConfig } from '@/types'

interface IEditBranchNameModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  isVisible: boolean
  onClose: () => void
}

export const EditBranchNameModal = ({
  branch,
  currentConfig,
  isVisible,
  onClose,
}: IEditBranchNameModal) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { refresh } = useBranch()
  const { addToast } = useToast()
  const [branchName, setBranchName] = useState(branch.name || '')
  const [repo, setRepo] = useState('')
  const [gitBranch, setGitBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [validationError, setValidationError] = useState<string | null>(null)

  // Initialize VCS fields from current config
  useEffect(() => {
    if (isVisible && currentConfig) {
      if (currentConfig.connected_github_vcs_config) {
        const vcs = currentConfig.connected_github_vcs_config
        setRepo(vcs.repo || '')
        setGitBranch(vcs.branch || 'main')
        setDirectory(vcs.directory || '.')
        setPathFilter(vcs.path_filter || '')
      } else if (currentConfig.public_git_vcs_config) {
        const vcs = currentConfig.public_git_vcs_config
        setRepo(vcs.repo || '')
        setGitBranch(vcs.branch || 'main')
        setDirectory(vcs.directory || '.')
        setPathFilter(vcs.path_filter || '')
      }
    }
  }, [isVisible, currentConfig])

  const formatError = (err: any): string => {
    if (!err) return ''
    if (typeof err === 'string') return err
    return (
      err.user_error ||
      err.error ||
      err.description ||
      err.message ||
      'An error occurred'
    )
  }

  const { mutate: handleSave, isPending: isSubmitting } = useMutation({
    mutationFn: async () => {
      if (!branchName.trim()) {
        throw new Error('Branch name cannot be empty')
      }

      // Step 1: Update branch name if changed
      if (branchName !== branch.name) {
        const { error: updateError } = await updateBranch({
          appId: app.id,
          branchId: branch.id || '',
          orgId: org.id,
          request: { name: branchName },
        })

        if (updateError) {
          throw new Error(formatError(updateError))
        }
      }

      // Step 2: Create new config with updated VCS settings (if config exists)
      if (currentConfig && repo && gitBranch) {
        const request: any = {}

        // Preserve VCS config type and update values
        if (currentConfig.connected_github_vcs_config) {
          request.connected_github_vcs_config = {
            vcs_connection_id:
              currentConfig.connected_github_vcs_config.vcs_connection_id || '',
            repo: repo.trim(),
            branch: gitBranch.trim(),
            directory: directory.trim(),
            path_filter: pathFilter.trim() || undefined,
          }
        } else if (currentConfig.public_git_vcs_config) {
          request.public_git_vcs_config = {
            repo: repo.trim(),
            branch: gitBranch.trim(),
            directory: directory.trim(),
            path_filter: pathFilter.trim() || undefined,
          }
        }

        // Preserve install groups if they exist
        if (
          currentConfig.install_groups &&
          currentConfig.install_groups.length > 0
        ) {
          request.install_groups = currentConfig.install_groups.map(
            (g, idx) => ({
              name: g.name,
              install_ids: g.install_ids || [],
              order: g.order ?? idx,
              max_parallel: g.max_parallel || 1,
              requires_approval: g.requires_approval || false,
              rollback_on_failure: g.rollback_on_failure || false,
            })
          )
        }

        const { error: configError } = await createBranchConfig({
          appId: app.id,
          branchId: branch.id || '',
          orgId: org.id,
          request,
        })

        if (configError) {
          throw new Error(formatError(configError))
        }
      }
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Branch updated successfully" theme="success">
          <Text>Updated branch: {branchName}</Text>
        </Toast>
      )
      setValidationError(null)
      refresh()
      onClose()
    },
    onError: (error: Error) => {
      setValidationError(error.message)
      addToast(
        <Toast heading="Branch update failed" theme="error">
          <Text>Failed to update branch.</Text>
          <Text>{error.message}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      isVisible={isVisible}
      onClose={onClose}
      heading="Edit Branch"
      size="3/4"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save Changes',
        onClick: () => handleSave(),
        disabled: isSubmitting || !branchName.trim(),
      }}
    >
      {validationError && (
        <Banner theme="error" className="mb-4">
          {validationError}
        </Banner>
      )}

      <div className="space-y-6">
        {/* Branch Name Section */}
        <div>
          <Text variant="h4" weight="strong" className="mb-4">
            Branch Name
          </Text>
          <label
            htmlFor="branch-name"
            className="block text-sm font-medium mb-2"
          >
            Name
          </label>
          <input
            id="branch-name"
            type="text"
            value={branchName}
            onChange={(e) => setBranchName(e.target.value)}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Enter branch name"
            disabled={isSubmitting}
            autoFocus
          />
          <Text variant="subtext" theme="neutral" className="mt-1">
            A descriptive name for this branch configuration
          </Text>
        </div>

        {/* VCS Configuration Section */}
        {currentConfig && (
          <div className="border-t pt-6">
            <Text variant="h4" weight="strong" className="mb-4">
              VCS Configuration
            </Text>
            <Banner theme="info" className="mb-4">
              Updating VCS configuration will create a new config version
            </Banner>

            <div className="space-y-4">
              <div>
                <label
                  htmlFor="repo"
                  className="block text-sm font-medium mb-2"
                >
                  Repository
                </label>
                <Input
                  id="repo"
                  type="text"
                  value={repo}
                  onChange={(e) => setRepo(e.target.value)}
                  placeholder="owner/repo"
                  disabled={isSubmitting}
                />
              </div>

              <div>
                <label
                  htmlFor="git-branch"
                  className="block text-sm font-medium mb-2"
                >
                  Branch
                </label>
                <Input
                  id="git-branch"
                  type="text"
                  value={gitBranch}
                  onChange={(e) => setGitBranch(e.target.value)}
                  placeholder="main"
                  disabled={isSubmitting}
                />
              </div>

              <div>
                <label
                  htmlFor="directory"
                  className="block text-sm font-medium mb-2"
                >
                  Directory
                </label>
                <Input
                  id="directory"
                  type="text"
                  value={directory}
                  onChange={(e) => setDirectory(e.target.value)}
                  placeholder="."
                  disabled={isSubmitting}
                />
                <Text variant="subtext" theme="neutral" className="mt-1">
                  Path to your application config (use &quot;.&quot; for root)
                </Text>
              </div>

              <div>
                <label
                  htmlFor="path-filter"
                  className="block text-sm font-medium mb-2"
                >
                  Path Filter (Optional)
                </label>
                <Input
                  id="path-filter"
                  type="text"
                  value={pathFilter}
                  onChange={(e) => setPathFilter(e.target.value)}
                  placeholder="^(src/|config/).*"
                  disabled={isSubmitting}
                />
                <Text variant="subtext" theme="neutral" className="mt-1">
                  Regex pattern to filter which file changes trigger workflow
                  runs
                </Text>
              </div>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
