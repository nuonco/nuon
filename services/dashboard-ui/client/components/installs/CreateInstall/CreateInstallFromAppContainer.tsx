import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import {
  buildCreateInstallBody,
  normalizeInstallPlatform,
} from '@/components/installs/forms/InstallForm'
import type { InstallFormValues } from '@/components/installs/forms/InstallForm'
import { useAuth } from '@/hooks/use-auth'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/posthog-analytics'
import {
  getAppConfigs,
  getAppConfig,
  getAppBranches,
  getBranchConfigs,
  createAppInstall,
  getAWSAccountConnections,
  getComponents,
  installNameTaken,
} from '@/lib'
import type {
  TApp,
  TAppBranch,
  TAppBranchInstallGroup,
  TAppConfig,
} from '@/types'
import { shouldDefaultStackOnly } from './app-install-readiness'
import { BranchStep } from './BranchStep'
import {
  GroupStep,
  concreteMatchLabels,
  type TGroupStepSelection,
} from './GroupStep'
import {
  CreateInstallFormFields,
  type ICreateFormTriggerState,
} from './CreateInstallFormFields'
import { FormSkeleton } from './FormSkeleton'

export type CreateInstallPhase = 'select-branch' | 'form' | 'pick-group'

export interface ICreateFromAppState extends ICreateFormTriggerState {
  isSubmitting: boolean
  phase: CreateInstallPhase
}

const noop = () => {}

export const pickCreateInstallConfig = (
  configs?: TAppConfig[],
  opts?: {
    requireUnbranched?: boolean
    requireActive?: boolean
    fallbackToFirst?: boolean
  }
) => {
  const selected = configs?.find((config) => {
    if (
      (opts?.requireActive && config.status !== 'active') ||
      (config.status && config.status !== 'active')
    )
      return false
    if (config.labels?.source === 'git-preview-run') return false
    if (opts?.requireUnbranched && config.app_branch_id) return false
    return true
  })
  return (
    selected ?? (opts?.fallbackToFirst === false ? undefined : configs?.[0])
  )
}

interface ICreateInstallFromAppContainer {
  app: TApp
  initialBranchId?: string
  onBack?: () => void
  onStateChange: (state: ICreateFromAppState) => void
  modalId?: string
}

export const CreateInstallFromAppContainer = ({
  app,
  initialBranchId,
  onBack,
  onStateChange,
  modalId,
}: ICreateInstallFromAppContainer) => {
  const { org } = useOrg()
  const { user } = useAuth()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const platform = app.runner_config?.app_runner_type
  const requireTargetAccount = !!org?.features?.['phone-home-auth']
  const awsConnectionsEnabled =
    platform === 'aws' && !!org?.features?.['aws-account-connections']

  const [fields, setFields] = useState<ICreateFormTriggerState>({
    canSubmit: false,
    submit: noop,
  })
  const [selectedBranch, setSelectedBranch] = useState<TAppBranch | null>(null)
  const [branchDecisionMade, setBranchDecisionMade] = useState(false)
  const [initialBranchApplied, setInitialBranchApplied] =
    useState(!initialBranchId)
  const [pendingFormValues, setPendingFormValues] =
    useState<InstallFormValues | null>(null)
  const [selectedGroup, setSelectedGroup] = useState<TGroupStepSelection>(null)

  const {
    data: branchList,
    isLoading: branchesLoading,
    isError: branchesError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches', org?.id, app.id],
    queryFn: () => getAppBranches({ appId: app.id, orgId: org?.id || '' }),
    enabled: !!org?.id && !!app.id,
  })
  const hasBranches = !branchesError && (branchList?.data ?? []).length > 0

  // Opening from a branch page preselects that branch and skips the picker.
  // Applied once so Back still returns to the picker.
  useEffect(() => {
    if (initialBranchApplied || branchesLoading) return
    const match = (branchList?.data ?? []).find(
      (branch) => branch.id === initialBranchId
    )
    if (match) {
      setSelectedBranch(match)
      setBranchDecisionMade(true)
    }
    setInitialBranchApplied(true)
  }, [initialBranchApplied, branchesLoading, branchList, initialBranchId])

  // Derive phase
  const phase: CreateInstallPhase = (() => {
    if (!initialBranchApplied) return 'select-branch'
    if (hasBranches && !branchDecisionMade) return 'select-branch'
    if (selectedBranch && pendingFormValues !== null) return 'pick-group'
    return 'form'
  })()

  const {
    data: unbranchedConfigs,
    isLoading: unbranchedConfigsLoading,
    error: unbranchedConfigsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && phase === 'form' && !selectedBranch,
  })

  const {
    data: branchConfigs,
    isLoading: branchConfigsLoading,
    error: branchConfigsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branch-app-configs', org?.id, app.id, selectedBranch?.id],
    queryFn: () =>
      getBranchConfigs({
        orgId: org.id,
        appId: app.id,
        branchId: selectedBranch!.id,
        limit: 100,
      }),
    enabled: !!org?.id && !!selectedBranch?.id && phase === 'form',
  })

  const configId = selectedBranch
    ? pickCreateInstallConfig(branchConfigs, {
        requireActive: true,
        fallbackToFirst: false,
      })?.id
    : pickCreateInstallConfig(unbranchedConfigs, {
        requireUnbranched: true,
        requireActive: true,
        fallbackToFirst: false,
      })?.id

  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app.id, configId],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: app.id,
        appConfigId: configId!,
        recurse: true,
      }),
    enabled: !!org?.id && !!configId && phase === 'form',
  })

  const {
    data: awsAccountConnections,
    isLoading: awsAccountConnectionsLoading,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['aws-account-connections', org?.id],
    queryFn: () => getAWSAccountConnections({ orgId: org.id }),
    enabled: !!org?.id && awsConnectionsEnabled && phase === 'form',
  })

  const componentIds = config?.component_ids ?? []
  const needsComponents = componentIds.length > 0
  const { data: componentsResult, isLoading: componentsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['components', org?.id, app.id, 'create-install-gate'],
    queryFn: () => getComponents({ orgId: org.id, appId: app.id, limit: 100 }),
    enabled: !!org?.id && !!app.id && needsComponents && !!config,
  })

  const {
    mutateAsync,
    isPending: isSubmitting,
    error: submitError,
  } = useMutation({
    mutationFn: (body: ReturnType<typeof buildCreateInstallBody>) =>
      createAppInstall({ appId: app.id, body, orgId: org?.id || '' }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_create',
        status: 'ok',
        user,
        props: {
          appId: app.id,
          installId: result.data.id,
        },
      })
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>
            Created {result.data?.name ?? 'install'}. Provisioning may take a
            few minutes.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      const suffix =
        result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
      removeModal(modalId)
      const workflowId = result.data.workflow_id
      navigate(
        workflowId
          ? `/${org?.id}/installs/${result.data.id}/workflows/${workflowId}${suffix}`
          : `/${org?.id}/installs/${result.data.id}/workflows${suffix}`
      )
    },
    onError: (err: any) => {
      trackEvent({
        event: 'install_create',
        status: 'error',
        user,
        props: {
          appId: app.id,
          err: err?.error,
        },
      })
    },
  })

  const isFormLoading =
    unbranchedConfigsLoading ||
    branchConfigsLoading ||
    configLoading ||
    (needsComponents && componentsLoading) ||
    (awsConnectionsEnabled && awsAccountConnectionsLoading)

  const missingBranchConfigError =
    phase === 'form' &&
    !!selectedBranch &&
    !branchConfigsLoading &&
    !branchConfigsError &&
    !configId
      ? { error: 'This branch has no active app config.' }
      : undefined
  const missingUnbranchedConfigError =
    phase === 'form' &&
    !selectedBranch &&
    !unbranchedConfigsLoading &&
    !unbranchedConfigsError &&
    !configId
      ? {
          error:
            'No active config from `nuon apps sync` was found. Sync the app before creating an install without a branch.',
        }
      : undefined
  const loadError =
    phase === 'form'
      ? unbranchedConfigsError ||
        branchConfigsError ||
        configError ||
        missingBranchConfigError ||
        missingUnbranchedConfigError
      : undefined
  const inputConfig = config?.input
    ? {
        ...config.input,
        input_groups: (config.input.input_groups || []).map((group) => ({
          ...group,
          app_inputs:
            config.input?.inputs?.filter(
              (input) => input.group_id === group.id
            ) || [],
        })),
      }
    : undefined

  const formReady =
    phase === 'form' && !isFormLoading && !loadError && !!inputConfig

  const validateName = useCallback(
    async (name: string) => {
      const trimmed = name.trim()
      if (!trimmed || !org?.id) return undefined
      try {
        const taken = await installNameTaken({
          appId: app.id,
          orgId: org.id,
          name: trimmed,
        })
        return taken ? `An install named "${trimmed}" already exists` : undefined
      } catch {
        // A failed lookup shouldn't block creation; the API still enforces
        // uniqueness on submit.
        return undefined
      }
    },
    [app.id, org?.id]
  )

  // Build merged labels for the pick-group → submit step
  const buildGroupLabels = (
    formValues: InstallFormValues,
    group: TGroupStepSelection
  ): Record<string, string> | undefined => {
    const formLabels: Record<string, string> = {}
    for (const { key, value } of formValues.labels ?? []) {
      const trimmed = key.trim()
      if (trimmed) formLabels[trimmed] = value.trim()
    }
    const groupLabels = group
      ? concreteMatchLabels(group as TAppBranchInstallGroup)
      : {}
    const merged = { ...formLabels, ...groupLabels }
    return Object.keys(merged).length > 0 ? merged : undefined
  }

  const submitFromPickGroup = () => {
    if (!pendingFormValues || !selectedBranch) return
    const base = buildCreateInstallBody(
      pendingFormValues,
      normalizeInstallPlatform(platform)
    )
    const mergedLabels = buildGroupLabels(pendingFormValues, selectedGroup)
    mutateAsync({
      ...base,
      labels: mergedLabels,
      app_branch_id: selectedBranch.id,
    })
  }

  useEffect(() => {
    if (phase === 'select-branch') {
      onStateChange({
        canSubmit: !!selectedBranch,
        submit: () => setBranchDecisionMade(true),
        isSubmitting: false,
        phase,
      })
      return
    }

    if (phase === 'pick-group') {
      onStateChange({
        canSubmit: !isSubmitting,
        submit: submitFromPickGroup,
        isSubmitting,
        phase,
      })
      return
    }

    // form phase
    if (selectedBranch) {
      // form submit → advance to pick-group, don't call API yet
      onStateChange({
        canSubmit: formReady ? fields.canSubmit : false,
        submit: fields.submit,
        isSubmitting: false,
        phase,
      })
      return
    }

    onStateChange({
      canSubmit: formReady ? fields.canSubmit : false,
      submit: fields.submit,
      isSubmitting,
      phase,
    })
  }, [
    formReady,
    fields,
    isSubmitting,
    phase,
    selectedBranch,
    selectedGroup,
    pendingFormValues,
    hasBranches,
    branchDecisionMade,
    onStateChange,
  ])

  const handleBack = () => {
    if (phase === 'pick-group') {
      setPendingFormValues(null)
      setSelectedGroup(null)
      return
    }
    if (phase === 'form' && branchDecisionMade) {
      setSelectedBranch(null)
      setBranchDecisionMade(false)
      return
    }
    onBack?.()
  }

  const showBack = phase !== 'select-branch' || !!onBack

  const backButton = showBack ? (
    <Button
      className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
      onClick={handleBack}
    >
      <Icon variant="CaretLeftIcon" weight="bold" />
      Back
    </Button>
  ) : null

  if (phase === 'select-branch') {
    return (
      <div className="flex flex-col gap-6">
        {onBack ? backButton : null}
        {branchesLoading ? (
          <FormSkeleton />
        ) : (
          <BranchStep
            branches={branchList?.data ?? []}
            selected={selectedBranch}
            onSelect={setSelectedBranch}
            onSkip={() => {
              setSelectedBranch(null)
              setBranchDecisionMade(true)
            }}
          />
        )}
      </div>
    )
  }

  if (phase === 'pick-group') {
    const branchConfig = selectedBranch?.configs?.at(0)
    const formLabels: Record<string, string> = {}
    for (const { key, value } of pendingFormValues?.labels ?? []) {
      const trimmed = key.trim()
      if (trimmed) formLabels[trimmed] = value.trim()
    }

    return (
      <div className="flex flex-col gap-6">
        {backButton}
        {submitError ? (
          <Banner theme="error">
            {(submitError as any)?.error ||
              (submitError as any)?.description ||
              'Unable to create install.'}
          </Banner>
        ) : null}
        {branchConfig ? (
          <GroupStep
            config={branchConfig}
            installLabels={formLabels}
            selected={selectedGroup}
            onSelect={setSelectedGroup}
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            This branch has no deployment plan configured. The install will be
            created on the branch without a group assignment.
          </Text>
        )}
      </div>
    )
  }

  // form phase
  return (
    <div className="flex flex-col gap-6">
      {backButton}

      {loadError ? (
        <Banner theme="error">
          {(loadError as any)?.error || 'Unable to load app configuration'}
        </Banner>
      ) : isFormLoading || !inputConfig ? (
        <FormSkeleton />
      ) : (
        <CreateInstallFormFields
          app={app}
          inputConfig={inputConfig}
          defaultStackOnly={shouldDefaultStackOnly(
            {
              ...app,
              app_configs: [{ component_ids: config?.component_ids ?? [] }],
            },
            needsComponents ? componentsResult?.data : []
          )}
          requireTargetAccount={requireTargetAccount}
          awsAccountConnections={
            awsConnectionsEnabled ? awsAccountConnections || [] : undefined
          }
          submitError={
            !selectedBranch && submitError
              ? ({
                  error:
                    (submitError as any).error ||
                    (submitError as any).description ||
                    'Unable to create install.',
                } as any)
              : null
          }
          validateName={validateName}
          onSubmit={
            selectedBranch
              ? (values) => {
                  // Don't call API yet — advance to group picker
                  setPendingFormValues(values)
                }
              : (values) =>
                  mutateAsync(
                    buildCreateInstallBody(
                      values,
                      normalizeInstallPlatform(platform)
                    )
                  )
          }
          onStateChange={setFields}
        />
      )}
    </div>
  )
}
