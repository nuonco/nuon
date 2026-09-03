import { useMemo, useState, useEffect } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { useVcsRepoBrowser } from '@/hooks/use-vcs-repo-browser'
import { getAppInstallsConfig, createAppInstallsConfig } from '@/lib'
import type { TAPIError } from '@/types'
import type { TCreateAppInstallsConfigBody } from '@/lib/ctl-api/apps/install-syncs/create-app-installs-config'
import { ManageInstallsConfig } from './ManageInstallsConfig'

type IManageInstallsConfigContainer = Omit<IModal, 'onSubmit'>

export const ManageInstallsConfigContainer = ({
  ...props
}: IManageInstallsConfigContainer) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const vcsConnections = org?.vcs_connections || []
  const [vcsConnectionId, setVcsConnectionId] = useState(
    vcsConnections[0]?.id || ''
  )

  const { data: currentConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs-config', org?.id, app?.id],
    queryFn: () => getAppInstallsConfig({ appId: app!.id, orgId: org!.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const vcsBrowser = useVcsRepoBrowser({
    orgId: org?.id ?? '',
    vcsConnectionId,
    enabled: !!vcsConnectionId,
  })

  const [didAutofill, setDidAutofill] = useState(false)

  useEffect(() => {
    if (didAutofill || !currentConfig) return

    if (currentConfig.vcs_connection_id) {
      setVcsConnectionId(currentConfig.vcs_connection_id)
    }

    const parts = currentConfig.repo.split('/')
    vcsBrowser.setSelectedRepo({
      id: 0,
      name: parts.length > 1 ? parts.slice(1).join('/') : currentConfig.repo,
      full_name: currentConfig.repo,
      private: currentConfig.vcs_type === 'connected',
      fork: false,
      html_url: `https://github.com/${currentConfig.repo}`,
      default_branch: currentConfig.branch,
      updated_at: '',
    })
    vcsBrowser.setSelectedBranch(currentConfig.branch)
    setDidAutofill(true)
  }, [
    currentConfig,
    didAutofill,
    vcsBrowser.setSelectedRepo,
    vcsBrowser.setSelectedBranch,
  ])

  const repos = useMemo(() => {
    const list = vcsBrowser.repos ?? []
    if (
      vcsBrowser.selectedRepo &&
      !list.some((r) => r.full_name === vcsBrowser.selectedRepo!.full_name)
    ) {
      return [vcsBrowser.selectedRepo, ...list]
    }
    return list
  }, [vcsBrowser.repos, vcsBrowser.selectedRepo])

  const {
    mutate,
    isPending: isSubmitting,
    error: submitError,
  } = useMutation({
    mutationFn: (body: TCreateAppInstallsConfigBody) =>
      createAppInstallsConfig({ appId: app!.id, body, orgId: org!.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-installs-config', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Installs config saved" theme="success">
          <Text>Updated the installs config source for {app?.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast heading="Save failed" theme="error">
          <Text>{error?.error || 'Unable to save installs config.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <ManageInstallsConfig
      currentConfig={currentConfig}
      vcsConnections={vcsConnections}
      repos={repos}
      branches={vcsBrowser.branches}
      loadingRepos={vcsBrowser.loadingRepos && !didAutofill}
      loadingBranches={vcsBrowser.loadingBranches}
      reposError={vcsBrowser.reposError}
      branchesError={vcsBrowser.branchesError}
      selectedVcsConnectionId={vcsConnectionId}
      onVcsConnectionChange={setVcsConnectionId}
      selectedRepo={vcsBrowser.selectedRepo}
      onRepoChange={vcsBrowser.setSelectedRepo}
      selectedBranch={vcsBrowser.selectedBranch}
      onBranchChange={vcsBrowser.setSelectedBranch}
      isSubmitting={isSubmitting}
      submitError={submitError}
      onSubmit={(body) => mutate(body)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const ManageInstallsConfigButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <ManageInstallsConfigContainer />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="GearIcon" size={16} />
      Manage
    </Button>
  )
}
