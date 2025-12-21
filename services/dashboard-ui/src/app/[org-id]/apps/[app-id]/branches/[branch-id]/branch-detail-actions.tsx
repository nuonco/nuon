'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import type { TAppBranch, TAppBranchConfig, TVCSConnection } from '@/types'
import { triggerBranchRun } from '@/lib'
import { EditBranchNamePanel } from './edit-branch-name-panel'
import { EditInstallGroupsPanel } from './edit-install-groups-panel'

interface IBranchDetailActions {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
}

export const BranchDetailActions = ({
  branch,
  currentConfig,
  appId,
  orgId,
}: IBranchDetailActions) => {
  const router = useRouter()
  const { addToast } = useToast()
  const [showEditName, setShowEditName] = useState(false)
  const [showEditGroups, setShowEditGroups] = useState(false)
  const [isTriggering, setIsTriggering] = useState(false)

  const handleTriggerRun = async () => {
    if (!currentConfig) {
      addToast(
        <Toast theme="error" heading="No configuration available">
          <Text>Create a config first before triggering a run.</Text>
        </Toast>
      )
      return
    }

    setIsTriggering(true)
    try {
      const { data, error } = await triggerBranchRun({
        appId,
        branchId: branch.id,
        orgId,
        request: {
          config_id: currentConfig.id,
          force: false,
        },
      })

      if (error) {
        const errorMessage =
          typeof error === 'string'
            ? error
            : error.user_error ||
              error.error ||
              error.description ||
              'Failed to trigger run'
        addToast(
          <Toast theme="error" heading="Failed to trigger run">
            <Text>{errorMessage}</Text>
          </Toast>
        )
      } else {
        addToast(
          <Toast theme="success" heading="Run triggered successfully">
            <Text>Your app branch run has been queued.</Text>
          </Toast>
        )
        // Refresh the page to show the new run
        router.refresh()
      }
    } catch (err) {
      addToast(
        <Toast theme="error" heading="Failed to trigger run">
          <Text>An unexpected error occurred.</Text>
        </Toast>
      )
    } finally {
      setIsTriggering(false)
    }
  }

  return (
    <>
      <div className="flex items-center gap-3">
        <Button
          onClick={() => setShowEditName(true)}
          variant="secondary"
          size="sm"
        >
          <Icon variant="Edit" size={16} />
          Edit
        </Button>

        <Button
          onClick={() => setShowEditGroups(true)}
          variant="secondary"
          size="sm"
        >
          <Icon variant="Edit" size={16} />
          Edit Installs
        </Button>

        <Button
          variant="primary"
          size="sm"
          disabled={!currentConfig || isTriggering}
          onClick={handleTriggerRun}
          title={
            !currentConfig
              ? 'Create a configuration first to trigger a run'
              : 'Trigger a new run with the current configuration'
          }
        >
          <Icon variant="Play" size={16} />
          {isTriggering ? 'Triggering...' : 'Trigger Run'}
        </Button>
      </div>

      <EditBranchNamePanel
        branch={branch}
        currentConfig={currentConfig}
        orgId={orgId}
        appId={appId}
        isVisible={showEditName}
        onClose={() => setShowEditName(false)}
      />

      <EditInstallGroupsPanel
        branch={branch}
        currentConfig={currentConfig}
        orgId={orgId}
        appId={appId}
        isVisible={showEditGroups}
        onClose={() => setShowEditGroups(false)}
      />
    </>
  )
}
