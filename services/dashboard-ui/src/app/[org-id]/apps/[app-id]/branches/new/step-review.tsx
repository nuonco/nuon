'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { IFormData } from './types'
import { mockVCSConnections, mockRepos, getMockInstalls } from './mock-data'

interface IStepReviewProps {
  formData: IFormData
  isSuccess?: boolean
  branchId?: string
  orgId?: string
  appId?: string
  onSubmit?: () => void
  isSubmitting?: boolean
}

export const StepReview = ({
  formData,
  isSuccess,
  branchId,
  orgId,
  appId,
  onSubmit,
  isSubmitting,
}: IStepReviewProps) => {
  const mockInstalls = getMockInstalls()
  const vcsConnection = mockVCSConnections.find(
    (v) => v.id === formData.vcsConnection
  )
  const repo = mockRepos.find((r) => r.id === formData.repo)

  const getInstallById = (id: string) => {
    return mockInstalls.find((i) => i.id === id)!
  }

  // Show success state if branch was created
  if (isSuccess && branchId) {
    return (
      <div className="space-y-6">
        <Card>
          <div className="flex flex-col items-center gap-4 py-12">
            <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/20 flex items-center justify-center">
              <Icon
                variant="CheckCircle"
                size={40}
                className="text-green-600"
              />
            </div>
            <div className="text-center space-y-2">
              <Text variant="h3" weight="strong">
                Branch Created Successfully!
              </Text>
              <Text variant="base" theme="neutral">
                Your app branch "{formData.branchName}" has been created with
                mock data.
              </Text>
              <Text
                variant="sm"
                theme="neutral"
                className="text-cool-grey-500 dark:text-cool-grey-400"
              >
                Configuration saved to localStorage for testing.
              </Text>
            </div>

            <div className="flex flex-col gap-3 mt-6 w-full max-w-md">
              <div className="p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-900">
                <div className="flex items-start gap-2">
                  <Icon
                    variant="Info"
                    size={16}
                    className="text-blue-600 mt-0.5"
                  />
                  <div className="flex-1">
                    <Text
                      variant="sm"
                      weight="strong"
                      className="text-blue-800 dark:text-blue-300 mb-1"
                    >
                      Testing with Mock Data
                    </Text>
                    <Text
                      variant="xs"
                      className="text-blue-700 dark:text-blue-400"
                    >
                      This branch uses mock data for UI testing. Check browser
                      DevTools → Application → Local Storage to see the saved
                      configuration.
                    </Text>
                  </div>
                </div>
              </div>

              <div className="flex gap-3 mt-2">
                <Button
                  onClick={() => {
                    if (orgId && appId) {
                      window.location.href = `/${orgId}/apps/${appId}/branches`
                    }
                  }}
                  variant="primary"
                  className="flex-1"
                >
                  View Branches
                </Button>
                <Button
                  onClick={() => window.location.reload()}
                  variant="secondary"
                  className="flex-1"
                >
                  Create Another
                </Button>
              </div>

              <div className="mt-4 p-3 bg-cool-grey-50 dark:bg-dark-grey-800 rounded border font-mono text-xs">
                <Text
                  variant="xs"
                  className="text-cool-grey-600 dark:text-cool-grey-400"
                >
                  localStorage key:
                </Text>
                <Text variant="xs" className="break-all">
                  app-branch-config-{branchId}
                </Text>
              </div>
            </div>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Summary Header */}
      <div className="flex items-center gap-2 mb-4">
        <Icon variant="CheckCircle" size={24} className="text-green-600" />
        <Text variant="lg" weight="strong">
          Review Configuration
        </Text>
      </div>

      {/* What Happens Next Info */}
      <Card>
        <div className="p-6 bg-blue-50 dark:bg-blue-950/20">
          <div className="flex items-start gap-3">
            <Icon variant="Info" size={20} className="text-blue-600 mt-1" />
            <div>
              <Text variant="sm" weight="strong" className="mb-2">
                What happens next?
              </Text>
              <ul className="list-disc list-inside space-y-1">
                <Text variant="sm" className="text-blue-800 dark:text-blue-300">
                  <li>Your branch configuration will be saved</li>
                  <li>
                    {formData.isManualOnly
                      ? 'Workflow runs can be triggered manually from the dashboard or CLI'
                      : 'Workflow runs will be triggered automatically on Git commits'}
                  </li>
                  <li>
                    Installs will deploy in the order you configured (
                    {formData.installGroups.length} group
                    {formData.installGroups.length !== 1 ? 's' : ''})
                  </li>
                  <li>
                    You can modify these settings anytime from the branch
                    settings page
                  </li>
                </Text>
              </ul>
            </div>
          </div>
        </div>
      </Card>

      {/* VCS Configuration Summary */}
      <Card>
        <div className="p-6 space-y-4">
          <div className="flex items-center gap-2 mb-4">
            <Icon variant="GitBranch" size={20} className="text-primary-600" />
            <Text variant="base" weight="strong">
              Branch Configuration
            </Text>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
              >
                Branch Name
              </Text>
              <Text variant="sm" weight="strong">
                {formData.branchName}
              </Text>
            </div>

            {formData.description && (
              <div className="md:col-span-2">
                <Text
                  variant="xs"
                  className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                >
                  Description
                </Text>
                <Text variant="sm">{formData.description}</Text>
              </div>
            )}

            <div>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
              >
                Workflow Trigger
              </Text>
              <Badge theme={formData.isManualOnly ? 'warn' : 'success'}>
                {formData.isManualOnly ? 'Manual Only' : 'Automatic'}
              </Badge>
            </div>
          </div>

          {!formData.isManualOnly && vcsConnection && repo && (
            <>
              <div className="border-t pt-4 mt-4">
                <Text
                  variant="sm"
                  weight="strong"
                  className="text-cool-grey-700 dark:text-cool-grey-300 mb-3"
                >
                  Version Control
                </Text>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Text
                      variant="xs"
                      className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                    >
                      VCS Connection
                    </Text>
                    <div className="flex items-center gap-2">
                      <Icon variant="GitHub" size={16} />
                      <Text variant="sm">{vcsConnection.name}</Text>
                    </div>
                  </div>

                  <div>
                    <Text
                      variant="xs"
                      className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                    >
                      Repository
                    </Text>
                    <Text variant="sm">{repo.name}</Text>
                  </div>

                  <div>
                    <Text
                      variant="xs"
                      className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                    >
                      Git Branch
                    </Text>
                    <div className="flex items-center gap-2">
                      <Icon variant="GitBranch" size={14} />
                      <Text variant="sm">{formData.gitBranch}</Text>
                    </div>
                  </div>

                  <div>
                    <Text
                      variant="xs"
                      className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                    >
                      Directory
                    </Text>
                    <Text variant="sm">{formData.directory}</Text>
                  </div>

                  {formData.pathFilter && (
                    <div className="md:col-span-2">
                      <Text
                        variant="xs"
                        className="text-cool-grey-600 dark:text-cool-grey-400 mb-1"
                      >
                        Path Filter
                      </Text>
                      <code className="text-sm bg-cool-grey-100 dark:bg-dark-grey-700 px-2 py-1 rounded">
                        {formData.pathFilter}
                      </code>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </div>
      </Card>

      {/* Deployment Order Summary */}
      <Card>
        <div className="p-6 space-y-4">
          <div className="flex items-center gap-2 mb-4">
            <Icon variant="Stack" size={20} className="text-primary-600" />
            <Text variant="base" weight="strong">
              Deployment Order
            </Text>
          </div>

          {formData.installGroups.length > 0 ? (
            <div className="space-y-4">
              {formData.installGroups.map((group, index) => {
                if (group.installIds.length === 0) return null

                return (
                  <div
                    key={group.id}
                    className="border rounded-lg p-4 bg-cool-grey-50 dark:bg-dark-grey-800"
                  >
                    <div className="flex items-center gap-2 mb-3">
                      <div className="w-6 h-6 rounded-full bg-primary-600 text-white flex items-center justify-center text-xs font-strong">
                        {index + 1}
                      </div>
                      <div className="flex-1">
                        <Text variant="sm" weight="strong">
                          {group.name}
                        </Text>
                        <Text
                          variant="xs"
                          className="text-cool-grey-600 dark:text-cool-grey-400"
                        >
                          {group.installIds.length} install
                          {group.installIds.length !== 1 ? 's' : ''} • Max{' '}
                          {group.maxParallel} parallel
                        </Text>
                      </div>
                    </div>

                    {/* Group Settings Summary */}
                    <div className="flex items-center gap-3 mb-3 flex-wrap">
                      {group.requiresApproval && (
                        <Badge size="sm" theme="warn">
                          Requires Approval
                        </Badge>
                      )}
                      {group.rollbackOnFailure && (
                        <Badge size="sm" theme="neutral">
                          Auto Rollback
                        </Badge>
                      )}
                    </div>

                    <div className="space-y-2">
                      {group.installIds.map((installId) => {
                        const install = getInstallById(installId)
                        return (
                          <div
                            key={installId}
                            className="flex items-center gap-3 p-2 bg-white dark:bg-dark-grey-900 rounded border"
                          >
                            <Icon variant="AWS" size={16} />
                            <div className="flex-1">
                              <Text variant="sm" weight="strong">
                                {install.name}
                              </Text>
                              <Text
                                variant="xs"
                                className="text-cool-grey-600 dark:text-cool-grey-400"
                              >
                                {install.region}
                              </Text>
                            </div>
                            <Badge
                              size="sm"
                              theme={
                                install.status === 'active'
                                  ? 'success'
                                  : 'neutral'
                              }
                            >
                              {install.status}
                            </Badge>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                )
              })}

              <div className="flex items-center gap-2 text-cool-grey-600 dark:text-cool-grey-400 mt-4 p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-900">
                <Icon variant="Info" size={16} />
                <Text variant="xs">
                  Total deployment steps:{' '}
                  {
                    formData.installGroups.filter(
                      (g) => g.installIds.length > 0
                    ).length
                  }{' '}
                  • Total installs:{' '}
                  {formData.installGroups.reduce(
                    (acc, g) => acc + g.installIds.length,
                    0
                  )}
                </Text>
              </div>
            </div>
          ) : (
            <Text
              variant="sm"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              No deployment groups configured
            </Text>
          )}
        </div>
      </Card>
    </div>
  )
}