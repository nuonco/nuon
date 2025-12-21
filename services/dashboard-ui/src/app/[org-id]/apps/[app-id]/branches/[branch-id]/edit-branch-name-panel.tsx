'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Modal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { updateBranch } from '@/lib'
import type { TAppBranch } from '@/types'

interface IEditBranchNamePanel {
  branch: TAppBranch
  orgId: string
  appId: string
  isVisible: boolean
  onClose: () => void
}

export const EditBranchNamePanel = ({
  branch,
  orgId,
  appId,
  isVisible,
  onClose,
}: IEditBranchNamePanel) => {
  const router = useRouter()
  const [branchName, setBranchName] = useState(branch.name || '')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSave = async () => {
    if (!branchName.trim()) {
      setError('Branch name cannot be empty')
      return
    }

    if (branchName === branch.name) {
      onClose()
      return
    }

    setIsSubmitting(true)
    setError(null)

    const { error: updateError } = await updateBranch({
      appId,
      branchId: branch.id || '',
      orgId,
      request: { name: branchName },
    })

    setIsSubmitting(false)

    if (updateError) {
      setError(
        typeof updateError === 'string'
          ? updateError
          : updateError.user_error ||
              updateError.error ||
              updateError.description ||
              'Failed to update branch name'
      )
    } else {
      router.refresh()
      onClose()
    }
  }

  return (
    <Modal
      isVisible={isVisible}
      onClose={onClose}
      heading="Edit Branch Name"
      size="half"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save Changes',
        onClick: handleSave,
        disabled: isSubmitting || !branchName.trim(),
      }}
    >
      <Banner theme="info" className="mb-4">
        To update VCS configuration or install groups, create a new configuration
        version.
      </Banner>

      {error && (
        <Banner theme="error" className="mb-4">
          {error}
        </Banner>
      )}

      <div className="space-y-4">
        <div>
          <label
            htmlFor="branch-name"
            className="block text-sm font-medium mb-2"
          >
            Branch Name
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
      </div>
    </Modal>
  )
}