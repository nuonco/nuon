import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getPreviewInstallCandidates,
  getPreviewSources,
  triggerBranchRun,
} from '@/lib'
import type {
  TAPIError,
  TAppBranch,
  TAppBranchConfig,
  TAppBranchPreviewConfig,
  TAppBranchRunPreviewMode,
  TAppBranchRunPreviewSource,
  TPreviewSourcePR,
} from '@/types'
import {
  previewDefaultsFromConfig,
} from '@/components/branches/shared/PreviewDefaultsEditor'
import {
  formatPreviewDefaultsSummary,
  isPreviewOverride,
  resolveInstallName,
  sortPreviewInstallCandidates,
} from '@/components/branches/shared/preview-run-utils'

export interface IPreviewBranchRunModal extends Omit<IModal, 'onSubmit'> {
  branchName: string
  branchDefaultsSummary: string
  hasOverride: boolean
  sourceTab: 'pr' | 'branch'
  onSourceTabChange: (tab: 'pr' | 'branch') => void
  prOptions: TPreviewSourcePR[]
  branchOptions: { name: string; sha?: string }[]
  selectedPR?: TPreviewSourcePR
  onSelectPR: (pr: TPreviewSourcePR) => void
  selectedBranch?: { name: string; sha?: string }
  onSelectBranch: (branch: { name: string; sha?: string }) => void
  installOptions: { value: string; label: string; description?: string }[]
  defaultInstallId: string
  selectedInstallId: string
  onSelectInstallId: (id: string) => void
  noInstallOptions: boolean
  mode: TAppBranchRunPreviewMode
  onModeChange: (mode: TAppBranchRunPreviewMode) => void
  loadingSources: boolean
  loadingInstalls: boolean
  isPending: boolean
  onConfirm: () => void
}

export const PreviewBranchRunModal = ({
  branchName,
  branchDefaultsSummary,
  hasOverride,
  sourceTab,
  onSourceTabChange,
  prOptions,
  branchOptions,
  selectedPR,
  onSelectPR,
  selectedBranch,
  onSelectBranch,
  installOptions,
  defaultInstallId,
  selectedInstallId,
  onSelectInstallId,
  noInstallOptions,
  mode,
  onModeChange,
  loadingSources,
  loadingInstalls,
  isPending,
  onConfirm,
  ...props
}: IPreviewBranchRunModal) => {
  const canSubmit =
    !isPending &&
    !loadingSources &&
    (sourceTab === 'pr' ? !!selectedPR : !!selectedBranch) &&
    (mode === 'build-only' || !!selectedInstallId)

  return (
    <Modal
      heading="Preview run"
      footerActions={
        <div className="flex flex-1 items-center gap-3 min-w-0 mr-auto">
          <Text variant="subtext" weight="strong" className="shrink-0">
            Mode
          </Text>
          <ToggleButton<TAppBranchRunPreviewMode>
            value={mode}
            onChange={onModeChange}
            options={[
              { value: 'plan-only', label: 'Plan only' },
              { value: 'apply', label: 'Apply' },
              { value: 'build-only', label: 'Build only' },
            ]}
          />
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? 'Triggering...' : 'Trigger preview',
        disabled: !canSubmit,
        onClick: onConfirm,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-5">
        <Text variant="subtext" theme="neutral">
          Branch <strong>{branchName}</strong>
        </Text>

        <Banner theme="info">
          <div className="flex flex-col gap-1">
            <Text variant="subtext" weight="strong">
              Branch defaults
            </Text>
            <Text variant="subtext">{branchDefaultsSummary}</Text>
            {hasOverride ? (
              <Text variant="subtext" theme="neutral">
                Overriding branch defaults for this run.
              </Text>
            ) : (
              <Text variant="subtext" theme="neutral">
                Using branch defaults.
              </Text>
            )}
          </div>
        </Banner>

        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Source
          </Text>
          <ToggleButton<'pr' | 'branch'>
            value={sourceTab}
            onChange={onSourceTabChange}
            options={[
              { value: 'pr', label: 'Pull request' },
              { value: 'branch', label: 'Branch' },
            ]}
          />
          {sourceTab === 'pr' ? (
            <>
              <Select
                options={prOptions.map((pr) => ({
                  value: String(pr.pr_number),
                  label: `#${pr.pr_number} · ${pr.title}`,
                  description: pr.head_ref,
                }))}
                value={selectedPR ? String(selectedPR.pr_number) : ''}
                onChange={(val) => {
                  const pr = prOptions.find((p) => String(p.pr_number) === val)
                  if (pr) onSelectPR(pr)
                }}
                placeholder={loadingSources ? 'Loading pull requests...' : 'Select a pull request'}
                disabled={isPending || loadingSources || prOptions.length === 0}
              />
              <Text variant="subtext" theme="neutral">
                Open pull requests onto this branch.
              </Text>
            </>
          ) : (
            <>
              <Select
                options={branchOptions.map((b) => ({
                  value: b.name,
                  label: b.name,
                }))}
                value={selectedBranch?.name ?? ''}
                onChange={(name) => {
                  const branch = branchOptions.find((b) => b.name === name)
                  if (branch) onSelectBranch(branch)
                }}
                placeholder={loadingSources ? 'Loading branches...' : 'Select a branch'}
                disabled={isPending || loadingSources || branchOptions.length === 0}
              />
              <Text variant="subtext" theme="neutral">
                Git branches available for preview.
              </Text>
            </>
          )}
        </div>

        {mode !== 'build-only' && (
          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Install
            </Text>
            <Select
              options={installOptions}
              value={selectedInstallId}
              onChange={onSelectInstallId}
              placeholder={loadingInstalls ? 'Loading installs...' : 'Select an install'}
              disabled={isPending || loadingInstalls || installOptions.length === 0}
              searchable={installOptions.length > 5}
              menuPlacement="bottom"
            />
            {noInstallOptions && !loadingInstalls ? (
              <Text variant="subtext" theme="neutral">
                No installs found for this app. Create an install first.
              </Text>
            ) : (
              <Text variant="subtext" theme="neutral">
                Installs on other branches are labeled in the list. Pre-filled from branch preview
                settings
                {defaultInstallId && selectedInstallId === defaultInstallId ? ' (default)' : ''}.
              </Text>
            )}
          </div>
        )}
      </div>
    </Modal>
  )
}

interface IPreviewBranchRunModalContainer extends Omit<IModal, 'onSubmit'> {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
  onSuccess?: () => void
}

export const PreviewBranchRunModalContainer = ({
  branch,
  currentConfig,
  appId,
  orgId,
  onSuccess,
  ...props
}: IPreviewBranchRunModalContainer) => {
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const { data: candidates, isLoading: loadingInstalls } = useQuery({
    queryKey: ['preview-install-candidates', orgId, appId, branch.id, currentConfig?.id],
    queryFn: () =>
      getPreviewInstallCandidates({
        orgId,
        appId,
        branchId: branch.id!,
        configId: currentConfig?.id,
      }),
    enabled: !!orgId && !!appId && !!branch.id,
  })

  const availableInstalls = useMemo(
    () => sortPreviewInstallCandidates(candidates?.installs ?? [], branch.id!),
    [candidates?.installs, branch.id]
  )

  const branchDefaults = useMemo(
    () => previewDefaultsFromConfig(currentConfig?.preview_config, availableInstalls),
    [currentConfig?.preview_config, availableInstalls]
  )

  const hasGithubVCS = !!(
    currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config
  )

  const branchDefaultsSummary = useMemo(
    () =>
      formatPreviewDefaultsSummary(currentConfig?.preview_config, availableInstalls, {
        includeGithub: hasGithubVCS,
      }),
    [currentConfig?.preview_config, availableInstalls, hasGithubVCS]
  )

  const [sourceTab, setSourceTab] = useState<'pr' | 'branch'>('pr')
  const [selectedPR, setSelectedPR] = useState<TPreviewSourcePR>()
  const [selectedBranch, setSelectedBranch] = useState<{ name: string; sha?: string }>()
  const [selectedInstallId, setSelectedInstallId] = useState('')
  const [mode, setMode] = useState<TAppBranchRunPreviewMode>('plan-only')
  const [defaultsSynced, setDefaultsSynced] = useState(false)

  useEffect(() => {
    if (loadingInstalls || defaultsSynced) return
    setMode(branchDefaults.mode)
    setSelectedInstallId((prev) => prev || branchDefaults.installId)
    setDefaultsSynced(true)
  }, [loadingInstalls, defaultsSynced, branchDefaults.installId, branchDefaults.mode])

  const { data: sources, isLoading: loadingSources } = useQuery({
    queryKey: ['preview-sources', orgId, appId, branch.id],
    queryFn: () =>
      getPreviewSources({ orgId, appId, branchId: branch.id! }),
    enabled: !!orgId && !!appId && !!branch.id,
  })

  useEffect(() => {
    const prs = sources?.pull_requests ?? []
    if (sourceTab === 'pr' && prs.length > 0 && !selectedPR) {
      setSelectedPR(prs[0])
    }
  }, [sources?.pull_requests, sourceTab, selectedPR])

  useEffect(() => {
    const branches = sources?.branches ?? []
    if (sourceTab === 'branch' && branches.length > 0 && !selectedBranch) {
      setSelectedBranch(branches[0])
    }
  }, [sources?.branches, sourceTab, selectedBranch])

  const installOptions = useMemo(
    () =>
      availableInstalls.map((i) => ({
        value: i.id,
        label:
          i.id === branchDefaults.installId
            ? `${i.name} (default)`
            : i.name,
        description:
          i.app_branch_id && i.app_branch_id !== branch.id
            ? `On branch ${i.app_branch?.name ?? 'another branch'}`
            : undefined,
      })),
    [availableInstalls, branchDefaults.installId, branch.id]
  )

  const hasOverride = isPreviewOverride(mode, selectedInstallId, branchDefaults)

  const { mutate, isPending } = useMutation({
    mutationFn: () => {
      const source: TAppBranchRunPreviewSource = sourceTab === 'pr' ? 'pr' : 'branch'
      const overrideMode = mode !== branchDefaults.mode ? mode : undefined
      const installId =
        mode !== 'build-only' && selectedInstallId ? selectedInstallId : undefined

      return triggerBranchRun({
        appId,
        branchId: branch.id!,
        orgId,
        request: {
          config_id: currentConfig?.id,
          preview_run: {
            source,
            pr_number: sourceTab === 'pr' ? selectedPR?.pr_number : undefined,
            git_ref: sourceTab === 'branch' ? selectedBranch?.name : undefined,
            head_sha:
              sourceTab === 'pr'
                ? selectedPR?.head_sha
                : selectedBranch?.sha,
            mode: overrideMode,
            install_id: installId,
          },
        },
      })
    },
    onSuccess: () => {
      addToast(
        <Toast theme="success" heading="Preview run triggered">
          <Text>Your preview run has been queued.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast theme="error" heading="Preview run failed">
          <Text>{error.error || 'Unable to trigger preview run.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <PreviewBranchRunModal
      branchName={branch.name || ''}
      branchDefaultsSummary={branchDefaultsSummary}
      hasOverride={hasOverride}
      sourceTab={sourceTab}
      onSourceTabChange={setSourceTab}
      prOptions={sources?.pull_requests ?? []}
      branchOptions={sources?.branches ?? []}
      selectedPR={selectedPR}
      onSelectPR={setSelectedPR}
      selectedBranch={selectedBranch}
      onSelectBranch={setSelectedBranch}
      installOptions={installOptions}
      defaultInstallId={branchDefaults.installId}
      selectedInstallId={selectedInstallId}
      onSelectInstallId={setSelectedInstallId}
      noInstallOptions={!loadingInstalls && installOptions.length === 0}
      mode={mode}
      onModeChange={setMode}
      loadingSources={loadingSources}
      loadingInstalls={loadingInstalls}
      isPending={isPending}
      onConfirm={() => mutate()}
      {...props}
    />
  )
}

export const quickPreviewFromDefaults = (
  config: TAppBranchPreviewConfig | undefined,
  modeOverride?: TAppBranchRunPreviewMode,
  installs?: { id: string; name?: string }[]
) => {
  const defaults = previewDefaultsFromConfig(config, installs as import('@/types').TInstall[])
  return {
    mode: modeOverride ?? defaults.mode,
    installId: defaults.installId,
    installName: defaults.installId
      ? resolveInstallName(defaults.installId, config, installs as import('@/types').TInstall[])
      : '',
  }
}
