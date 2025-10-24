'use client'

import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import type { THelmK8sChangeAction, TTerraformChangeAction } from '@/types'
import { cn } from '@/utils/classnames'
import {
  getTerraformActionBorderColor,
  getHelmActionBorderColor,
} from './diff-style-utils'

// Terraform plan actions
const TERRAFORM_ACTIONS: {
  value: TTerraformChangeAction
  label: string
}[] = [
  { value: 'create', label: 'Create' },
  { value: 'update', label: 'Update' },
  { value: 'delete', label: 'Delete' },
  { value: 'replace', label: 'Replace' },
  { value: 'read', label: 'Read' },
  { value: 'noop', label: 'No-op' },
]

// Other helm/k8s actions
const HELM_DIFF_ACTIONS: {
  value: string
  label: string
}[] = [
  { value: 'added', label: 'Added' },
  { value: 'destroyed', label: 'Destroyed' },
  { value: 'changed', label: 'Changed' },
]

interface IDiffFilter {
  title: string
  selectedActions: Set<string>
  onInputToggle: (action: string) => void
  onButtonClick: (action: string) => void
  onReset: () => void
  selectedCount: number
  totalCount: number
  searchValue: string
  onSearchChange: (value: string) => void
  searchPlaceholder: string
  diffType?: 'terraform' | 'helm-k8s'
}

export function DiffFilter({
  title,
  selectedActions,
  onInputToggle,
  onButtonClick,
  onReset,
  selectedCount,
  totalCount,
  searchValue,
  onSearchChange,
  searchPlaceholder,
  diffType = 'terraform',
}: IDiffFilter) {
  const getButtonText = (action: string) => {
    // If only this action is selected, show "Reset"
    if (selectedActions.size === 1 && selectedActions.has(action)) {
      return 'Reset'
    }
    // Otherwise, show "Only"
    return 'Only'
  }

  // Choose which actions to display based on diffType
  const actionOptions =
    diffType === 'terraform' ? TERRAFORM_ACTIONS : HELM_DIFF_ACTIONS

  return (
    <div className="px-4 sm:px-6 py-4 border-b bg-cool-grey-25 dark:bg-dark-grey-800 flex items-center justify-between gap-4">
      <SearchInput
        placeholder={searchPlaceholder}
        value={searchValue}
        onChange={onSearchChange}
      />
      <div className="flex items-center gap-6">
        <Text variant="subtext" theme="neutral">
          {selectedCount} of {totalCount} {title}
        </Text>
      </div>

      <Dropdown
        buttonClassName="!p-1"
        buttonText={`Filter ${title} (${selectedActions.size})`}
        className="ml-auto"
        icon={<Icon variant="FadersHorizontal" />}
        id={`${title}-fitler`}
        alignment="right"
        variant="ghost"
        size="sm"
      >
        <Menu className="min-w-56">
          {actionOptions.map((option) => (
            <div key={option.value} className="flex items-center space-x-2">
              <input
                type="checkbox"
                id={`${title.toLowerCase().replace(' ', '-')}-action-${option.value}`}
                checked={selectedActions.has(option.value)}
                onChange={() => onInputToggle(option.value)}
                className="focus:ring-primary-500 focus:border-primary-500 accent-primary-600"
              />
              <Button
                className="!p-1 flex items-center justify-between group w-full"
                variant="ghost"
                size="sm"
                onClick={() => onButtonClick(option.value)}
              >
                <span className="flex items-center gap-1.5 h-full">
                  <span
                    className={cn('inline-block border-l-2 h-full', {
                      [getTerraformActionBorderColor(
                        option.value as TTerraformChangeAction
                      )]: diffType === 'terraform',
                      [getHelmActionBorderColor(
                        option.value as THelmK8sChangeAction
                      )]: diffType === 'helm-k8s',
                    })}
                  />
                  {option.label}
                </span>
                <span className="hidden group-hover:inline-flex text-xs">
                  {getButtonText(option.value)}
                </span>
              </Button>
            </div>
          ))}

          <hr />
          <Button
            className="w-full !p-1 shrink-0"
            variant="ghost"
            size="sm"
            onClick={onReset}
          >
            Reset
          </Button>
        </Menu>
      </Dropdown>
    </div>
  )
}
