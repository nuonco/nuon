import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import {
  buildCreateInstallBody,
  normalizeInstallPlatform,
} from '@/components/installs/forms/InstallForm'
import { useAuth } from '@/hooks/use-auth'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/posthog-analytics'
import {
  getAppConfigs,
  getAppConfig,
  getAppBranches,
  createAppInstall,
  getAWSAccountConnections,
} from '@/lib'
import type { TApp } from '@/types'
import { BranchConnectionStep } from './BranchConnectionStep'
import {
  CreateInstallFormFields,
  type ICreateFormTriggerState,
} from './CreateInstallFormFields'
import { FormSkeleton } from './FormSkeleton'

export type CreateInstallPhase = 'form' | 'branches'

export interface ICreateFromAppState extends ICreateFormTriggerState {
  isSubmitting: boolean
  phase: CreateInstallPhase
}

const noop = () => {}

interface ICreateInstallFromAppContainer {
  app: TApp
  onBack?: () => void
  onStateChange: (state: ICreateFromAppState) => void
  modalId?: string
}

export const CreateInstallFromAppContainer = ({
  app,
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
  const [createdInstall, setCreatedInstall] = useState<{
    id: string
    workflowId?: string
    suffix: string
  } | null>(null)
  const [fields, setFields] = useState<ICreateFormTriggerState>({
    canSubmit: false,
    submit: noop,
  })

  const {
    data: configs,
    isLoading: configsLoading,
    error: configsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id,
  })

  const configId = configs?.[0]?.id

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
    enabled: !!org?.id && !!configId,
  })

  const { data: awsAccountConnections, isLoading: awsAccountConnectionsLoading } =
    useQuery({
      placeholderData: keepPreviousData,
      queryKey: ['aws-account-connections', org?.id],
      queryFn: () => getAWSAccountConnections({ orgId: org.id }),
      enabled: !!org?.id && awsConnectionsEnabled,
    })

  const { data: branchList, isSuccess: branchesLoaded } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches', org?.id, app.id],
    queryFn: () => getAppBranches({ appId: app.id, orgId: org?.id || '' }),
    enabled: !!org?.id && !!app.id,
  })
  const hasBranches = (branchList?.data ?? []).length > 0

  const { mutateAsync, isPending: isSubmitting, error: submitError } =
    useMutation({
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
        setCreatedInstall({
          id: result.data.id,
          workflowId: result.data.workflow_id,
          suffix,
        })
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

  const isLoading =
    configsLoading ||
    configLoading ||
    (awsConnectionsEnabled && awsAccountConnectionsLoading)
  const loadError = configsError || configError
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

  const showBranches = !!createdInstall && branchesLoaded && hasBranches
  const phase: CreateInstallPhase = showBranches ? 'branches' : 'form'
  const working = isSubmitting || (!!createdInstall && !branchesLoaded)
  const ready = !isLoading && !loadError && !!inputConfig && !createdInstall

  const navigateToInstall = () => {
    removeModal(modalId)
    if (!createdInstall) return
    const { id, workflowId, suffix } = createdInstall
    navigate(
      workflowId
        ? `/${org?.id}/installs/${id}/workflows/${workflowId}${suffix}`
        : `/${org?.id}/installs/${id}/workflows${suffix}`
    )
  }

  useEffect(() => {
    onStateChange({
      canSubmit: ready ? fields.canSubmit : false,
      submit: fields.submit,
      isSubmitting: working,
      phase,
    })
  }, [ready, fields, working, phase, onStateChange])

  useEffect(() => {
    if (!createdInstall || !branchesLoaded || hasBranches) return
    removeModal(modalId)
    const { id, workflowId, suffix } = createdInstall
    navigate(
      workflowId
        ? `/${org?.id}/installs/${id}/workflows/${workflowId}${suffix}`
        : `/${org?.id}/installs/${id}/workflows${suffix}`
    )
  }, [
    createdInstall,
    branchesLoaded,
    hasBranches,
    removeModal,
    navigate,
    modalId,
    org?.id,
  ])

  if (phase === 'branches') {
    return (
      <BranchConnectionStep
        appId={app.id}
        installId={createdInstall!.id}
        onDone={navigateToInstall}
        onSkip={() => removeModal(modalId)}
      />
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {onBack ? (
        <Button
          className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
          onClick={onBack}
        >
          <Icon variant="CaretLeftIcon" weight="bold" />
          Back
        </Button>
      ) : null}

      {loadError ? (
        <Banner theme="error">
          {(loadError as any)?.error || 'Unable to load app configuration'}
        </Banner>
      ) : isLoading || !inputConfig ? (
        <FormSkeleton />
      ) : (
        <CreateInstallFormFields
          app={app}
          inputConfig={inputConfig}
          requireTargetAccount={requireTargetAccount}
          awsAccountConnections={
            awsConnectionsEnabled ? awsAccountConnections || [] : undefined
          }
          submitError={
            submitError
              ? ({
                  error:
                    (submitError as any).error ||
                    (submitError as any).description ||
                    'Unable to create install.',
                } as any)
              : null
          }
          onSubmit={(values) =>
            mutateAsync(
              buildCreateInstallBody(values, normalizeInstallPlatform(platform))
            )
          }
          onStateChange={setFields}
        />
      )}
    </div>
  )
}
