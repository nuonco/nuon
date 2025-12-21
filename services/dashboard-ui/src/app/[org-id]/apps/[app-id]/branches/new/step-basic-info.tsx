'use client'

import { Input } from '@/components/old/Input'
import { Text } from '@/components/common/Text'
import { IFormData } from './types'

interface IStepBasicInfoProps {
  formData: IFormData
  updateFormData: (updates: Partial<IFormData>) => void
}

export const StepBasicInfo = ({
  formData,
  updateFormData,
}: IStepBasicInfoProps) => {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      // Trigger Next button if name is filled
      if (formData.branchName.trim()) {
        const nextButton = document.querySelector('[data-wizard-next]') as HTMLButtonElement
        nextButton?.click()
      }
    }
  }

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
            A unique name for this branch configuration (e.g., main, production, staging)
          </Text>
        </label>
        <Input
          type="text"
          value={formData.branchName}
          onChange={(e) => updateFormData({ branchName: e.target.value })}
          onKeyDown={handleKeyDown}
          placeholder="main"
          required
        />
      </div>

      {/* Description */}
      <div className="space-y-2">
        <label className="block">
          <Text variant="sm" weight="strong">
            Description
          </Text>
          <Text
            variant="xs"
            className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
          >
            Optional description to help identify this branch's purpose
          </Text>
        </label>
        <textarea
          value={formData.description || ''}
          onChange={(e) => updateFormData({ description: e.target.value })}
          placeholder="Production deployment configuration for main branch"
          rows={3}
          className="w-full px-3 py-2 border border-cool-grey-300 dark:border-dark-grey-600 rounded-md bg-white dark:bg-dark-grey-800 text-cool-grey-900 dark:text-cool-grey-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
        />
      </div>
    </div>
  )
}