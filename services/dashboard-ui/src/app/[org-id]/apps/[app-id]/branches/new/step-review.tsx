'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { IFormData, mockVCSConnections, mockRepos, mockInstalls } from './page'

interface IStepReviewProps {
  formData: IFormData
}

export const StepReview = ({ formData }: IStepReviewProps) => {
  const vcsConnection = mockVCSConnections.find(
    (v) => v.id === formData.vcsConnection
  )
  const repo = mockRepos.find((r) => r.id === formData.repo)

  const getInstallById = (id: string) => {
    return mockInstalls.find((i) => i.id === id)!
  }

  return (
    <div className="space-y-6">
      {/* Summary Header */}
      <div className="flex items-center gap-2">
        <Icon variant="CheckCircle" size={24} className="text-green-600" />
        <Text variant="lg" weight="strong">
          Review Configuration
        </Text>
      </div>

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

          {formData.deploymentGroups.length > 0 ? (
            <div className="space-y-4">
              {formData.deploymentGroups.map((group, index) => {
                if (group.length === 0) return null

                return (
                  <div
                    key={index}
                    className="border rounded-lg p-4 bg-cool-grey-50 dark:bg-dark-grey-800"
                  >
                    <div className="flex items-center gap-2 mb-3">
                      <div className="w-6 h-6 rounded-full bg-primary-600 text-white flex items-center justify-center text-xs font-strong">
                        {index + 1}
                      </div>
                      <Text variant="sm" weight="strong">
                        Group {index + 1}
                      </Text>
                      <Text
                        variant="xs"
                        className="text-cool-grey-600 dark:text-cool-grey-400"
                      >
                        • {group.length} install{group.length !== 1 ? 's' : ''}{' '}
                        • Parallel deployment
                      </Text>
                    </div>

                    <div className="space-y-2">
                      {group.map((installId) => {
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
                  {formData.deploymentGroups.filter((g) => g.length > 0).length}{' '}
                  • Total installs:{' '}
                  {formData.deploymentGroups.reduce(
                    (acc, g) => acc + g.length,
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
