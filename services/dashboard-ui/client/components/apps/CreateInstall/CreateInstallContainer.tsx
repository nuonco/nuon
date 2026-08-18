import { useState } from 'react'
import { useNavigate } from 'react-router'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import {
  CreateInstallFormFields,
  FormSkeleton,
  type ICreateFormTriggerState,
} from '@/components/installs/CreateInstall'
import {
  buildCreateInstallBody,
  normalizeInstallPlatform,
} from '@/components/installs/forms/InstallForm'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getAppConfigs,
  getAppConfig,
  createAppInstall,
  getAWSAccountConnections,
} from '@/lib'
import { toSentenceCase } from '@/utils/string-utils'
import { CreateInstallButton as CreateInstallButtonComponent } from './CreateInstall'

const noop = () => {}

const CreateInstallModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { app } = useApp()
  const [trigger, setTrigger] = useState<ICreateFormTriggerState>({
    canSubmit: false,
    submit: noop,
  })
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const platform = app?.runner_config?.app_runner_type
  const requireTargetAccount = !!org?.features?.['phone-home-auth']
  const awsConnectionsEnabled =
    platform === 'aws' && !!org?.features?.['aws-account-connections']

  const {
    data: configs,
    isLoading: configsLoading,
    error: configsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const configId = configs?.[0]?.id

  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app?.id, configId],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: app.id,
        appConfigId: configId!,
        recurse: true,
      }),
    enabled: !!configId,
  })

  const { data: awsAccountConnections, isLoading: awsAccountConnectionsLoading } =
    useQuery({
      placeholderData: keepPreviousData,
      queryKey: ['aws-account-connections', org?.id],
      queryFn: () => getAWSAccountConnections({ orgId: org.id }),
      enabled: !!org?.id && awsConnectionsEnabled,
    })

  const { mutateAsync, isPending: isSubmitting, error: submitError } =
    useMutation({
      mutationFn: (body: ReturnType<typeof buildCreateInstallBody>) =>
        createAppInstall({ appId: app?.id || '', body, orgId: org?.id || '' }),
      onSuccess: (result) => {
        addToast(
          <Toast heading="Install created" theme="success">
            <Text>Install created.</Text>
          </Toast>
        )
        queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
        queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
        removeModal(props.modalId)
        const workflowId = result.data.workflow_id
        const suffix =
          result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
        navigate(
          workflowId
            ? `/${org?.id}/installs/${result.data.id}/workflows/${workflowId}${suffix}`
            : `/${org?.id}/installs/${result.data.id}/workflows${suffix}`
        )
      },
    })

  const isLoading =
    configsLoading ||
    configLoading ||
    (awsConnectionsEnabled && awsAccountConnectionsLoading)
  const loadError =
    configsError || configError || (!!configs && configs.length === 0)
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

  const ready = !isLoading && !loadError && !!inputConfig

  return (
    <Modal
      {...props}
      size="xl"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      showFooter={ready}
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="CubeIcon" size="24" />
          Create install
        </Text>
      }
      primaryActionTrigger={
        ready
          ? {
              children: isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" />
                  Creating install
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Icon variant="PlusIcon" />
                  Create install
                </span>
              ),
              disabled: !trigger.canSubmit || isSubmitting,
              onClick: () => trigger.submit(),
              variant: 'primary',
            }
          : undefined
      }
    >
      {loadError ? (
        <Banner theme="error">
          {(configsError as any)?.error ||
            (configError as any)?.error ||
            'Unable to load app configuration'}
        </Banner>
      ) : !ready || !inputConfig ? (
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
                  error: toSentenceCase(
                    (submitError as any).error ||
                      (submitError as any).description ||
                      'Unable to create install.'
                  ),
                } as any)
              : null
          }
          onSubmit={(values) =>
            mutateAsync(
              buildCreateInstallBody(values, normalizeInstallPlatform(platform))
            )
          }
          onStateChange={setTrigger}
        />
      )}
    </Modal>
  )
}

export const CreateInstallButtonContainer = ({
  onClick: _onClick,
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <CreateInstallButtonComponent
      onClick={() => addModal(<CreateInstallModalContainer />)}
      {...props}
    />
  )
}
