'use client'

import { useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Banner } from '@/components/common/Banner'
import { Input, CheckboxInput } from '@/components/old/Input'
import { Dropdown } from '@/components/old/Dropdown'
import { RadioInput } from '@/components/old/Input'
import { IFormData, IInstallGroup } from './types'
import { getMockInstalls, getDefaultGroupTemplates } from './mock-data'
import { InstallCard } from './install-card'
import { cn } from '@/utils/classnames'

interface IStepInstallGroupsProps {
  formData: IFormData
  updateFormData: (updates: Partial<IFormData>) => void
}

export const StepInstallGroups = ({
  formData,
  updateFormData,
}: IStepInstallGroupsProps) => {
  const [selectedTemplate, setSelectedTemplate] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const mockInstalls = getMockInstalls()
  const templates = getDefaultGroupTemplates()

  const handleApplyTemplate = (templateId: string) => {
    const template = templates.find((t) => t.id === templateId)
    if (!template) return

    const newGroups: IInstallGroup[] = template.groups.map((g, index) => ({
      id: `group-${Date.now()}-${index}`,
      name: g.name,
      installIds: [],
      order: index,
      requiresApproval: g.requiresApproval,
      rollbackOnFailure: g.rollbackOnFailure,
      maxParallel: g.maxParallel,
    }))

    updateFormData({
      installGroups: newGroups,
    })
    setSelectedTemplate(templateId)
  }

  const handleAddGroup = () => {
    const newGroup: IInstallGroup = {
      id: `group-${Date.now()}`,
      name: `Group ${formData.installGroups.length + 1}`,
      installIds: [],
      order: formData.installGroups.length,
      requiresApproval: false,
      rollbackOnFailure: true,
      maxParallel: 5,
    }

    updateFormData({
      installGroups: [...formData.installGroups, newGroup],
    })
  }

  const handleDeleteGroup = (groupId: string) => {
    const group = formData.installGroups.find((g) => g.id === groupId)
    if (!group) return

    const newGroups = formData.installGroups.filter((g) => g.id !== groupId)

    updateFormData({
      installGroups: newGroups,
      ungroupedInstalls: [...formData.ungroupedInstalls, ...group.installIds],
    })
  }

  const handleUpdateGroup = (groupId: string, updates: Partial<IInstallGroup>) => {
    const newGroups = formData.installGroups.map((g) =>
      g.id === groupId ? { ...g, ...updates } : g
    )
    updateFormData({ installGroups: newGroups })
  }

  const handleMoveGroupUp = (groupId: string) => {
    const index = formData.installGroups.findIndex((g) => g.id === groupId)
    if (index <= 0) return

    const newGroups = [...formData.installGroups]
    const temp = newGroups[index]
    newGroups[index] = newGroups[index - 1]
    newGroups[index - 1] = temp

    // Update order
    newGroups.forEach((g, i) => {
      g.order = i
    })

    updateFormData({ installGroups: newGroups })
  }

  const handleMoveGroupDown = (groupId: string) => {
    const index = formData.installGroups.findIndex((g) => g.id === groupId)
    if (index < 0 || index >= formData.installGroups.length - 1) return

    const newGroups = [...formData.installGroups]
    const temp = newGroups[index]
    newGroups[index] = newGroups[index + 1]
    newGroups[index + 1] = temp

    // Update order
    newGroups.forEach((g, i) => {
      g.order = i
    })

    updateFormData({ installGroups: newGroups })
  }

  const handleMoveToGroup = (installId: string, groupId: string) => {
    const newGroups = formData.installGroups.map((g) =>
      g.id === groupId
        ? { ...g, installIds: [...g.installIds, installId] }
        : g
    )
    const newUngrouped = formData.ungroupedInstalls.filter((id) => id !== installId)

    updateFormData({
      installGroups: newGroups,
      ungroupedInstalls: newUngrouped,
    })
  }

  const handleMoveToUngrouped = (installId: string, groupId: string) => {
    const newGroups = formData.installGroups.map((g) =>
      g.id === groupId
        ? { ...g, installIds: g.installIds.filter((id) => id !== installId) }
        : g
    )

    updateFormData({
      installGroups: newGroups,
      ungroupedInstalls: [...formData.ungroupedInstalls, installId],
    })
  }

  const getInstallById = (id: string) => {
    return mockInstalls.find((i) => i.id === id)!
  }

  const hasInstallsInGroups = formData.installGroups.some(
    (group) => group.installIds.length > 0
  )

  const filteredUngroupedInstalls = formData.ungroupedInstalls.filter((id) => {
    const install = getInstallById(id)
    return install.name.toLowerCase().includes(searchQuery.toLowerCase())
  })

  return (
    <div className="space-y-6">
      {/* Template Selector */}
      {formData.installGroups.length === 0 && (
        <div className="space-y-4">
          <Banner theme="info">
            <div className="space-y-2">
              <Text variant="sm" weight="strong">
                Choose a Deployment Strategy
              </Text>
              <Text variant="sm">
                Select a template to get started, or create custom groups from scratch.
              </Text>
            </div>
          </Banner>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {templates.map((template) => (
              <button
                key={template.id}
                onClick={() => handleApplyTemplate(template.id)}
                className="text-left p-4 border-2 rounded-lg hover:border-primary-500 hover:bg-primary-50/30 dark:hover:bg-primary-950/20 transition-colors"
              >
                <Text variant="sm" weight="strong" className="mb-2">
                  {template.name}
                </Text>
                <Text
                  variant="xs"
                  className="text-cool-grey-600 dark:text-cool-grey-400 mb-3"
                >
                  {template.description}
                </Text>
                {template.groups.length > 0 && (
                  <div className="flex items-center gap-2 flex-wrap">
                    {template.groups.map((g, i) => (
                      <div
                        key={i}
                        className="text-xs px-2 py-1 bg-cool-grey-100 dark:bg-dark-grey-700 rounded"
                      >
                        {g.name}
                      </div>
                    ))}
                  </div>
                )}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Deployment Groups */}
      {formData.installGroups.length > 0 && (
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

          <div className="space-y-4">
            {formData.installGroups.map((group, index) => (
              <GroupCard
                key={group.id}
                group={group}
                index={index}
                totalGroups={formData.installGroups.length}
                onUpdate={(updates) => handleUpdateGroup(group.id, updates)}
                onDelete={() => handleDeleteGroup(group.id)}
                onMoveUp={() => handleMoveGroupUp(group.id)}
                onMoveDown={() => handleMoveGroupDown(group.id)}
                onRemoveInstall={(installId) =>
                  handleMoveToUngrouped(installId, group.id)
                }
                getInstallById={getInstallById}
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
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Icon variant="Package" size={20} className="text-cool-grey-600" />
              <Text variant="base" weight="strong">
                Available Installs
              </Text>
              <Text
                variant="sm"
                className="text-cool-grey-600 dark:text-cool-grey-400"
              >
                ({filteredUngroupedInstalls.length})
              </Text>
            </div>
          </div>

          {/* Search */}
          <Input
            type="text"
            placeholder="Search installs..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />

          {formData.installGroups.length === 0 ? (
            <Banner theme="warn">
              <Text variant="sm">
                Create at least one deployment group to organize your installs.
              </Text>
            </Banner>
          ) : (
            <div className="grid grid-cols-1 gap-2">
              {filteredUngroupedInstalls.map((installId) => {
                const install = getInstallById(installId)
                return (
                  <InstallCard
                    key={installId}
                    install={install}
                    availableGroups={formData.installGroups.map((g) => g.id)}
                    availableGroupNames={formData.installGroups.map((g) => g.name)}
                    onMoveToGroup={(groupId) => handleMoveToGroup(installId, groupId)}
                  />
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* Validation Message */}
      {!hasInstallsInGroups && formData.installGroups.length > 0 && (
        <Banner theme="warn">
          <Text variant="sm">
            Add at least one install to a deployment group to proceed.
          </Text>
        </Banner>
      )}
    </div>
  )
}

// GroupCard Component
interface IGroupCardProps {
  group: IInstallGroup
  index: number
  totalGroups: number
  onUpdate: (updates: Partial<IInstallGroup>) => void
  onDelete: () => void
  onMoveUp: () => void
  onMoveDown: () => void
  onRemoveInstall: (installId: string) => void
  getInstallById: (id: string) => any
}

const GroupCard = ({
  group,
  index,
  totalGroups,
  onUpdate,
  onDelete,
  onMoveUp,
  onMoveDown,
  onRemoveInstall,
  getInstallById,
}: IGroupCardProps) => {
  const [showSettings, setShowSettings] = useState(false)

  return (
    <div
      className={cn(
        'border-2 rounded-lg p-4',
        group.installIds.length > 0
          ? 'border-primary-300 dark:border-primary-700 bg-primary-50/30 dark:bg-primary-950/20'
          : 'border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50/50 dark:bg-dark-grey-800/50'
      )}
    >
      {/* Group Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3 flex-1">
          <div
            className={cn(
              'w-8 h-8 rounded-full flex items-center justify-center text-sm font-strong flex-shrink-0',
              group.installIds.length > 0
                ? 'bg-primary-600 text-white'
                : 'bg-cool-grey-300 dark:bg-dark-grey-600 text-cool-grey-700 dark:text-cool-grey-300'
            )}
          >
            {index + 1}
          </div>

          <div className="flex-1 min-w-0">
            <Input
              type="text"
              value={group.name}
              onChange={(e) => onUpdate({ name: e.target.value })}
              placeholder="Group name"
              className="font-strong"
            />
            <Text
              variant="xs"
              className="text-cool-grey-600 dark:text-cool-grey-400 mt-1"
            >
              {group.installIds.length} install{group.installIds.length !== 1 ? 's' : ''}{' '}
              • Deploy in parallel
            </Text>
          </div>
        </div>

        <div className="flex items-center gap-1 flex-shrink-0">
          <Button
            size="sm"
            variant="ghost"
            onClick={onMoveUp}
            disabled={index === 0}
            title="Move up"
          >
            <Icon variant="CaretUp" size={16} />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={onMoveDown}
            disabled={index === totalGroups - 1}
            title="Move down"
          >
            <Icon variant="CaretDown" size={16} />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={onDelete}
            title="Delete group"
          >
            <Icon variant="Trash" size={16} />
          </Button>
        </div>
      </div>

      {/* Installs in Group */}
      {group.installIds.length > 0 ? (
        <div className="space-y-2 mb-4">
          {group.installIds.map((installId) => {
            const install = getInstallById(installId)
            return (
              <InstallCard
                key={installId}
                install={install}
                isInGroup
                onMoveToUngrouped={() => onRemoveInstall(installId)}
              />
            )
          })}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-6 mb-4 border-2 border-dashed rounded border-cool-grey-300 dark:border-dark-grey-600">
          <Icon
            variant="Package"
            size={32}
            className="text-cool-grey-400 dark:text-cool-grey-600 mb-2"
          />
          <Text
            variant="sm"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            No installs in this group
          </Text>
        </div>
      )}

      {/* Settings Toggle */}
      <button
        type="button"
        onClick={() => setShowSettings(!showSettings)}
        className="flex items-center gap-2 text-primary-600 dark:text-primary-400 hover:underline mb-2"
      >
        <Icon
          variant={showSettings ? 'CaretDown' : 'CaretRight'}
          size={16}
        />
        <Text variant="sm" weight="strong">
          Group Settings
        </Text>
      </button>

      {/* Settings Panel */}
      {showSettings && (
        <div className="space-y-4 pl-6 pt-2 border-l-2 border-primary-300 dark:border-primary-700">
          <CheckboxInput
            name={`${group.id}-approval`}
            checked={group.requiresApproval}
            onChange={(e) => onUpdate({ requiresApproval: e.target.checked })}
            labelText="Require manual approval before deployment"
            labelClassName="!px-0 !py-0 hover:bg-transparent dark:hover:bg-transparent"
          />

          <CheckboxInput
            name={`${group.id}-rollback`}
            checked={group.rollbackOnFailure}
            onChange={(e) => onUpdate({ rollbackOnFailure: e.target.checked })}
            labelText="Automatically rollback on failure"
            labelClassName="!px-0 !py-0 hover:bg-transparent dark:hover:bg-transparent"
          />

          <div>
            <Text variant="sm" weight="strong" className="mb-2">
              Max Parallel Installs
            </Text>
            <Input
              type="number"
              min={1}
              max={20}
              value={group.maxParallel}
              onChange={(e) =>
                onUpdate({ maxParallel: parseInt(e.target.value) || 1 })
              }
              className="w-24"
            />
            <Text
              variant="xs"
              className="text-cool-grey-600 dark:text-cool-grey-400 mt-1"
            >
              Maximum number of installs to deploy simultaneously
            </Text>
          </div>
        </div>
      )}
    </div>
  )
}