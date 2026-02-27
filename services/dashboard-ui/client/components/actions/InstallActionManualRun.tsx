import { useNavigate } from 'react-router'
import React, { useRef, useState, type FormEvent } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { RoleSelector } from '@/components/common/form/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { runAction } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TAction, TActionConfig } from '@/types'

interface IInstallActionManualRunModal extends Omit<IModal, 'heading'> {
  action: TAction
  actionConfigId: string
}

function normalizeEnvVars(steps: TActionConfig['steps']) {
  const envVars = steps.reduce((acc, step) => {
    const keys = Object.keys(step?.env_vars || {})
    if (keys?.length) {
      keys.forEach((key) => {
        if (!acc[key]) {
          acc[key] = step?.env_vars[key]
        }
      })
    }
    return acc
  }, {} as Record<string, string>)

  return envVars
}

export const InstallActionManualRunModal = ({
  action,
  actionConfigId,
  ...props
}: IInstallActionManualRunModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const config = action?.configs?.[0]
  const envVars = normalizeEnvVars(config?.steps || [])

  const [customVars, setCustomVars] = useState<number[]>([])
  const [selectedRole, setSelectedRole] = useState<string>('')
  const formRef = useRef<HTMLFormElement>(null)

  const { error, isPending: isLoading, mutate, data } = useMutation({
    mutationFn: (body: Parameters<typeof runAction>[0]['body'] & { role?: string }) =>
      runAction({ body, installId: install.id, orgId: org.id }),
    onSuccess: (result) => {
      trackEvent({
        event: 'action_run',
        user,
        status: 'ok',
        props: {
          orgId: org?.id,
          installId: install?.id,
          actionConfigId: actionConfigId,
        },
      })
      addToast(
        <Toast heading="Action workflow started" theme="success">
          <Text>The action workflow {action?.name} has been started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'action_run',
        user,
        status: 'error',
        props: {
          orgId: org?.id,
          installId: install?.id,
          actionConfigId: actionConfigId,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Failed to run action workflow" theme="error">
          <Text>There was an error running {action?.name}.</Text>
          <Text variant="subtext">{err?.error || 'Unknown error occurred'}</Text>
        </Toast>
      )
    },
  })

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const overwrite = Object.fromEntries(new FormData(e.currentTarget))

    const vars = Object.keys(overwrite).reduce((acc, key) => {
      if (overwrite[key] == envVars[key]) return acc

      const customKey = key.split(':')
      if (customKey?.at(0) === 'custom' && customKey?.at(2) === 'name') {
        const varName = overwrite[key]
        const varValue = overwrite[`${customKey?.at(0)}:${customKey?.at(1)}:value`]
        if (typeof varName === 'string' && typeof varValue === 'string') {
          acc[varName] = varValue
        }
      } else if (customKey?.at(0) === 'custom' && customKey?.at(2) === 'value') {
        return acc
      } else {
        const value = overwrite[key]
        if (typeof value === 'string') {
          acc[key] = value
        }
      }

      return acc
    }, {} as Record<string, string>)

    mutate({
      action_workflow_config_id: actionConfigId,
      ...(vars && Object.keys(vars)?.length > 0 && { run_env_vars: vars }),
      ...(selectedRole && { role: selectedRole }),
    })
  }

  const handlePrimaryAction = () => {
    formRef.current?.requestSubmit()
  }

  return (
    <Modal
      heading={`Run action ${action?.name}?`}
      size="half"
      primaryActionTrigger={{
        children: isLoading ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Running action...
          </>
        ) : (
          <>
            <Icon variant="Play" />
            Run action
          </>
        ),
        disabled: isLoading,
        onClick: handlePrimaryAction,
        variant: 'primary',
      }}
      {...props}
    >

      <form ref={formRef} onSubmit={handleSubmit} className="flex flex-col gap-4">
        <RoleSelector
          installId={install?.id}
          operationType="trigger"
          principalType="action"
          value={selectedRole}
          onChange={(e) => setSelectedRole(e.target.value)}
          name="role"
        />
        <Expand
          id="action-env-vars"
          heading={<Text weight="strong">Edit environment variables</Text>}
          className="border rounded-md"
        >
          <div className="p-4 border-t flex flex-col gap-4">
            <Text variant="subtext">
              Edit or add custom environment variables for this manual action
              workflow run.
            </Text>

            {Object.keys(envVars).length > 0 && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {Object.keys(envVars).map((envVar) => (
                  <label key={envVar} className="flex flex-col gap-1">
                    <Text variant="label" weight="strong">
                      {envVar}
                    </Text>
                    <input
                      className="px-3 py-2 text-base rounded-md border bg-black/5 dark:bg-white/5 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600"
                      required
                      defaultValue={envVars[envVar]}
                      name={envVar}
                      type="text"
                    />
                  </label>
                ))}
              </div>
            )}

            {customVars.length > 0 && (
              <div className="flex flex-col gap-2">
                {customVars.map((cv) => (
                  <fieldset
                    key={cv}
                    className="grid grid-cols-1 md:grid-cols-2 gap-2 py-2 border-t relative"
                  >
                    <legend className="text-base font-medium pr-2 mb-2 flex items-center justify-between">
                      <span>Custom env var {cv + 1}</span>
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => {
                          setCustomVars((vars) => vars.filter((v) => v !== cv))
                        }}
                        className="ml-2 !p-2"
                        size="sm"
                        aria-label={`Remove custom env var ${cv + 1}`}
                      >
                        <Icon variant="X" size="12" />
                      </Button>
                    </legend>
                    <label className="flex flex-col gap-1">
                      <Text variant="label" weight="strong">
                        Name
                      </Text>
                      <input
                        className="px-3 py-2 text-base rounded-md border bg-black/5 dark:bg-white/5 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600"
                        required
                        name={`custom:${cv}:name`}
                        type="text"
                      />
                    </label>
                    <label className="flex flex-col gap-1">
                      <Text variant="label" weight="strong">
                        Value
                      </Text>
                      <input
                        className="px-3 py-2 text-base rounded-md border bg-black/5 dark:bg-white/5 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600"
                        required
                        name={`custom:${cv}:value`}
                        type="text"
                      />
                    </label>
                  </fieldset>
                ))}
              </div>
            )}

            <div>
              <Button
                type="button"
                variant="ghost"
                onClick={() => {
                  setCustomVars((vars) => [...vars, vars.length])
                }}
              >
                <Icon variant="Plus" />
                Add environment variable
              </Button>
            </div>
          </div>
        </Expand>

      </form>
    </Modal>
  )
}

export const InstallActionManualRunButton = ({
  action,
  actionConfigId,
  children = 'Run action',
  ...props
}: {
  action: TAction
  actionConfigId: string
} & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <InstallActionManualRunModal action={action} actionConfigId={actionConfigId} />
  )

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {children}
    </Button>
  )
}
