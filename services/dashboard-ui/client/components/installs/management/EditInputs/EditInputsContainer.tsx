import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateInstall, updateInstallInputs } from '@/lib'
import type { TAppConfig, TAppInputConfig } from '@/types'
import { EditInstallModal, type IEditInputsUpdatePayload } from './EditInputs'

interface IEditInputs {
  showNameField?: boolean
}

const nestInputsUnderGroups = (
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
): TAppInputConfig['input_groups'] =>
  groups
    ? groups.map((group) => ({
        ...group,
        app_inputs:
          inputs?.filter((input) => input.group_id === group.id) || [],
      }))
    : []

export const EditInputsFormModalContainer = ({
  showNameField,
  ...props
}: IEditInputs & Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { appConfig: config, isLoading, error } = useInstallAppConfig()

  const {
    mutateAsync: updateNameAsync,
    isPending: isUpdatingName,
    error: nameError,
  } = useMutation({
    mutationFn: (name: string) =>
      updateInstall({ body: { name }, installId: install.id, orgId: org.id }),
    onSuccess: (_data, name) => {
      addToast(
        <Toast heading="Install renamed" theme="success">
          <Text>Install renamed to {name}.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['install'] })
      removeModal(props.modalId)
    },
  })

  const {
    mutateAsync: updateInputsAsync,
    isPending: isUpdatingInputs,
    error: inputsError,
  } = useMutation({
    mutationFn: async (payload: IEditInputsUpdatePayload) => {
      if (payload.name) {
        await updateInstall({
          body: { name: payload.name },
          installId: install.id,
          orgId: org.id,
        })
      }
      return updateInstallInputs({
        installId: install.id,
        orgId: org.id,
        body: {
          inputs: payload.inputs,
          deploy_dependents: payload.deployDependents,
          ...(payload.inputsOnly && { inputs_only: true }),
          ...(payload.role && { role: payload.role }),
        },
      })
    },
    onSuccess: (result) => {
      addToast(
        <Toast heading="Install updated" theme="success">
          <Text>{install.name} has been updated.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install'] })
      removeModal(props.modalId)
      const workflowId = result?.data?.workflow_id
      navigate(
        workflowId
          ? `/${org.id}/installs/${install.id}/workflows/${workflowId}`
          : `/${org.id}/installs/${install.id}/workflows`
      )
    },
  })

  const isSubmitting = isUpdatingName || isUpdatingInputs
  const submitError = (nameError || inputsError) as any

  const inputConfig: TAppInputConfig | undefined = config?.input
    ? {
        ...config.input,
        input_groups: nestInputsUnderGroups(
          config.input.input_groups,
          config.input.inputs
        ),
      }
    : undefined

  if (isLoading || error || !inputConfig) {
    return (
      <Modal
        {...props}
        size="lg"
        className="!max-h-[80vh]"
        childrenClassName="overflow-y-auto"
        heading={
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="PencilSimpleLineIcon" size="24" />
            {showNameField ? 'Edit install' : 'Edit install inputs'}
          </Text>
        }
      >
        {error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to load app configuration'}
          </Banner>
        ) : (
          <div className="flex flex-col gap-6 max-w-3xl">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start"
              >
                <div className="flex flex-col gap-1">
                  <Skeleton width="140px" height="16px" />
                  <Skeleton width="180px" height="14px" />
                </div>
                <Skeleton width="100%" height="40px" />
              </div>
            ))}
          </div>
        )}
      </Modal>
    )
  }

  return (
    <EditInstallModal
      {...props}
      install={install}
      inputConfig={inputConfig}
      showNameField={showNameField}
      isSubmitting={isSubmitting}
      submitError={submitError}
      onSubmitName={updateNameAsync}
      onSubmitInputs={updateInputsAsync}
    />
  )
}

const MANAGED_BY_CONFIG_TIP = 'Managed by config. Disable config sync to edit.'

export const EditInputsButton = ({
  showNameField,
  ...props
}: IEditInputs & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  if (isManagedByConfig) {
    return (
      <Button
        disabled
        tooltipProps={{
          tipContent: MANAGED_BY_CONFIG_TIP,
          position: 'left',
          tipContentClassName:
            '!whitespace-normal !w-auto max-w-[200px] text-xs',
          className: 'w-full',
        }}
        {...props}
      >
        {props?.isMenuButton ? null : <Icon variant="PencilSimpleLineIcon" />}
        {showNameField ? 'Edit install' : 'Edit inputs'}
        {props?.isMenuButton ? <Icon variant="PencilSimpleLineIcon" /> : null}
      </Button>
    )
  }

  return (
    <Button
      onClick={() =>
        addModal(<EditInputsFormModalContainer showNameField={showNameField} />)
      }
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="PencilSimpleLineIcon" />}
      {showNameField ? 'Edit install' : 'Edit inputs'}
      {props?.isMenuButton ? <Icon variant="PencilSimpleLineIcon" /> : null}
    </Button>
  )
}
