'use client'

import { useState } from 'react'
import { Input, CheckboxInput } from '@/components/old/Input'
import { Dropdown } from '@/components/old/Dropdown'
import { RadioInput } from '@/components/old/Input'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { IFormData, mockVCSConnections, mockRepos, mockBranches } from './page'

interface IStepVCSConfigProps {
  formData: IFormData
  updateFormData: (updates: Partial<IFormData>) => void
}

export const StepVCSConfig = ({
  formData,
  updateFormData,
}: IStepVCSConfigProps) => {
  const [showMoreOptions, setShowMoreOptions] = useState(false)

  return (
    <div className="space-y-6">
      {/* Branch Name */}
      <div className="space-y-2">
        <label className="block">
          <Text variant="sm" weight="strong">
            Branch Name <span className="text-red-600">*</span>
          </Text>
          <Text
            variant="xs"
            className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
          >
            A unique name for this branch configuration
          </Text>
        </label>
        <Input
          type="text"
          value={formData.branchName}
          onChange={(e) => updateFormData({ branchName: e.target.value })}
          placeholder="main"
          required
        />
      </div>

      {/* Manual Mode Toggle */}
      <div className="border rounded-lg p-4 bg-cool-grey-50 dark:bg-dark-grey-800">
        <CheckboxInput
          name="isManualOnly"
          checked={formData.isManualOnly}
          onChange={(e) => updateFormData({ isManualOnly: e.target.checked })}
          labelText="Allow only manual workflow runs"
          labelClassName="!px-0 !py-0 hover:bg-transparent dark:hover:bg-transparent"
          labelTextClassName="!text-sm !font-strong"
        />
        <Text
          variant="xs"
          className="text-cool-grey-600 dark:text-cool-grey-400 mt-2 ml-7"
        >
          When enabled, builds will only be triggered manually via the dashboard
          or CLI
        </Text>
      </div>

      {/* VCS Configuration Section */}
      {!formData.isManualOnly && (
        <div className="space-y-6 pt-4">
          <div className="flex items-center gap-2 mb-4">
            <Icon variant="GitBranch" size={20} className="text-primary-600" />
            <Text variant="base" weight="strong">
              Version Control Configuration
            </Text>
          </div>

          {/* VCS Connection */}
          <div className="space-y-2">
            <label className="block">
              <Text variant="sm" weight="strong">
                VCS Connection <span className="text-red-600">*</span>
              </Text>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
              >
                Select the VCS connection to use
              </Text>
            </label>
            <Dropdown
              id="vcs-connection"
              text={
                formData.vcsConnection
                  ? mockVCSConnections.find(
                      (v) => v.id === formData.vcsConnection
                    )?.name
                  : 'Select VCS connection'
              }
              variant="secondary"
              isFullWidth
            >
              {mockVCSConnections.map((vcs) => (
                <RadioInput
                  key={vcs.id}
                  name="vcsConnection"
                  value={vcs.id}
                  checked={formData.vcsConnection === vcs.id}
                  onChange={(e) =>
                    updateFormData({ vcsConnection: e.target.value })
                  }
                  labelText={
                    <div className="flex items-center gap-2">
                      <Icon variant="GitHub" size={16} />
                      {vcs.name}
                    </div>
                  }
                />
              ))}
            </Dropdown>
          </div>

          {/* Repository */}
          <div className="space-y-2">
            <label className="block">
              <Text variant="sm" weight="strong">
                Repository <span className="text-red-600">*</span>
              </Text>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
              >
                Select the repository to watch
              </Text>
            </label>
            <Dropdown
              id="repository"
              text={
                formData.repo
                  ? mockRepos.find((r) => r.id === formData.repo)?.name
                  : 'Select repository'
              }
              variant="secondary"
              isFullWidth
              disabled={!formData.vcsConnection}
            >
              {mockRepos.map((repo) => (
                <RadioInput
                  key={repo.id}
                  name="repository"
                  value={repo.id}
                  checked={formData.repo === repo.id}
                  onChange={(e) => updateFormData({ repo: e.target.value })}
                  labelText={
                    <div className="flex items-center gap-2">
                      {repo.private && <Icon variant="Lock" size={14} />}
                      {repo.name}
                    </div>
                  }
                />
              ))}
            </Dropdown>
          </div>

          {/* Git Branch */}
          <div className="space-y-2">
            <label className="block">
              <Text variant="sm" weight="strong">
                Git Branch <span className="text-red-600">*</span>
              </Text>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
              >
                The branch to monitor for changes
              </Text>
            </label>
            <Dropdown
              id="git-branch"
              text={formData.gitBranch || 'Select branch'}
              variant="secondary"
              isFullWidth
              disabled={!formData.repo}
            >
              {mockBranches.map((branch) => (
                <RadioInput
                  key={branch}
                  name="gitBranch"
                  value={branch}
                  checked={formData.gitBranch === branch}
                  onChange={(e) =>
                    updateFormData({ gitBranch: e.target.value })
                  }
                  labelText={
                    <div className="flex items-center gap-2">
                      <Icon variant="GitBranch" size={14} />
                      {branch}
                    </div>
                  }
                />
              ))}
            </Dropdown>
          </div>

          {/* Directory */}
          <div className="space-y-2">
            <label className="block">
              <Text variant="sm" weight="strong">
                Directory
              </Text>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
              >
                Path to the application directory in the repository
              </Text>
            </label>
            <Input
              type="text"
              value={formData.directory}
              onChange={(e) => updateFormData({ directory: e.target.value })}
              placeholder="."
              disabled={!formData.repo}
            />
          </div>

          {/* More Options Collapsible */}
          <div className="border-t pt-4">
            <button
              type="button"
              onClick={() => setShowMoreOptions(!showMoreOptions)}
              className="flex items-center gap-2 text-primary-600 dark:text-primary-400 hover:underline"
            >
              <Icon
                variant={showMoreOptions ? 'CaretDown' : 'CaretRight'}
                size={16}
              />
              <Text variant="sm" weight="strong">
                Show more options
              </Text>
            </button>

            {showMoreOptions && (
              <div className="mt-4 space-y-4 pl-6">
                {/* Path Filter */}
                <div className="space-y-2">
                  <label className="block">
                    <Text variant="sm" weight="strong">
                      Path Filter (regex)
                    </Text>
                    <Text
                      variant="xs"
                      className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
                    >
                      Only trigger builds when files matching this pattern
                      change
                    </Text>
                  </label>
                  <Input
                    type="text"
                    value={formData.pathFilter}
                    onChange={(e) =>
                      updateFormData({ pathFilter: e.target.value })
                    }
                    placeholder="^(src/|config/)"
                    disabled={!formData.repo}
                  />
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
