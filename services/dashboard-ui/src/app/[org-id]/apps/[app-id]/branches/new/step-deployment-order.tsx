'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Banner } from '@/components/common/Banner'
import { IFormData, mockInstalls } from './page'
import { DeploymentGroup } from './deployment-group'
import { InstallCard } from './install-card'

interface IStepDeploymentOrderProps {
  formData: IFormData
  updateFormData: (updates: Partial<IFormData>) => void
}

export const StepDeploymentOrder = ({
  formData,
  updateFormData,
}: IStepDeploymentOrderProps) => {
  const handleAddGroup = () => {
    updateFormData({
      deploymentGroups: [...formData.deploymentGroups, []],
    })
  }

  const handleDeleteGroup = (groupIndex: number) => {
    const newGroups = [...formData.deploymentGroups]
    const removedInstalls = newGroups[groupIndex]
    newGroups.splice(groupIndex, 1)

    updateFormData({
      deploymentGroups: newGroups,
      ungroupedInstalls: [...formData.ungroupedInstalls, ...removedInstalls],
    })
  }

  const handleMoveToGroup = (installId: string, groupIndex: number) => {
    const newGroups = [...formData.deploymentGroups]
    const newUngrouped = formData.ungroupedInstalls.filter(
      (id) => id !== installId
    )

    newGroups[groupIndex] = [...newGroups[groupIndex], installId]

    updateFormData({
      deploymentGroups: newGroups,
      ungroupedInstalls: newUngrouped,
    })
  }

  const handleMoveToUngrouped = (installId: string, groupIndex: number) => {
    const newGroups = [...formData.deploymentGroups]
    newGroups[groupIndex] = newGroups[groupIndex].filter(
      (id) => id !== installId
    )

    updateFormData({
      deploymentGroups: newGroups,
      ungroupedInstalls: [...formData.ungroupedInstalls, installId],
    })
  }

  const getInstallById = (id: string) => {
    return mockInstalls.find((i) => i.id === id)!
  }

  const hasInstallsInGroups = formData.deploymentGroups.some(
    (group) => group.length > 0
  )

  return (
    <div className="space-y-6">
      {/* Instructions Banner */}
      <Banner theme="info">
        <div className="space-y-2">
          <Text variant="sm" weight="strong">
            Configure Deployment Order
          </Text>
          <Text variant="sm">
            Organize your installs into deployment groups. Groups deploy
            sequentially (1 → 2 → 3), while installs within a group deploy in
            parallel.
          </Text>
          <ul className="list-disc list-inside space-y-1 mt-2">
            <li className="text-sm">
              Use the arrow buttons to move installs into groups
            </li>
            <li className="text-sm">
              Create multiple groups for staged deployments
            </li>
            <li className="text-sm">
              Group deployment waits for all installs to complete before
              proceeding
            </li>
          </ul>
        </div>
      </Banner>

      {/* Deployment Groups */}
      {formData.deploymentGroups.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Icon variant="Stack" size={20} className="text-primary-600" />
              <Text variant="base" weight="strong">
                Deployment Groups
              </Text>
            </div>
            <Text
              variant="sm"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              Sequential execution →
            </Text>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {formData.deploymentGroups.map((group, index) => (
              <DeploymentGroup
                key={index}
                groupIndex={index}
                installs={group.map(getInstallById)}
                onRemoveInstall={(installId) =>
                  handleMoveToUngrouped(installId, index)
                }
                onDeleteGroup={() => handleDeleteGroup(index)}
              />
            ))}
          </div>
        </div>
      )}

      {/* Add Group Button */}
      <div className="flex justify-center">
        <Button variant="secondary" onClick={handleAddGroup}>
          <Icon variant="Plus" size={16} />
          Create New Group
        </Button>
      </div>

      {/* Ungrouped Installs */}
      {formData.ungroupedInstalls.length > 0 && (
        <div className="space-y-4 pt-4 border-t">
          <div className="flex items-center gap-2">
            <Icon variant="Package" size={20} className="text-cool-grey-600" />
            <Text variant="base" weight="strong">
              Ungrouped Installs
            </Text>
            <Text
              variant="sm"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              ({formData.ungroupedInstalls.length})
            </Text>
          </div>

          {formData.deploymentGroups.length === 0 ? (
            <Banner theme="warn">
              <Text variant="sm">
                Create at least one deployment group to organize your installs.
              </Text>
            </Banner>
          ) : (
            <div className="grid grid-cols-1 gap-2">
              {formData.ungroupedInstalls.map((installId) => {
                const install = getInstallById(installId)
                return (
                  <InstallCard
                    key={installId}
                    install={install}
                    availableGroups={formData.deploymentGroups.map((_, i) => i)}
                    onMoveToGroup={(groupIndex) =>
                      handleMoveToGroup(installId, groupIndex)
                    }
                  />
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* Validation Message */}
      {!hasInstallsInGroups && formData.deploymentGroups.length > 0 && (
        <Banner theme="warn">
          <Text variant="sm">
            Add at least one install to a deployment group to proceed.
          </Text>
        </Banner>
      )}
    </div>
  )
}
