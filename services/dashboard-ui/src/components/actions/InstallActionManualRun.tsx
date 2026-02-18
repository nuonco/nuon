'use client'

import { usePathname, useRouter } from 'next/navigation'
import React, { useRef, useState, type FormEvent } from 'react'
import { runAction } from '@/actions/installs/run-action'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
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
  const router = useRouter()
  const path = usePathname()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const config = action?.configs?.[0]
  const envVars = normalizeEnvVars(config?.steps || [])

  const [customVars, setCustomVars] = useState<number[]>([])
  const formRef = useRef<HTMLFormElement>(null)

  const { data, error, isLoading, execute, headers } = useServerAction({
    action: runAction,
  })

  useServerActionToast({
    data,
    error,
    successHeading: 'Action workflow started',
    successContent: (
      <Text>The action workflow {action?.name} has been started.</Text>
    ),
    errorHeading: 'Failed to run action workflow',
    errorContent: (
      <>
        <Text>There was an error running {action?.name}.</Text>
        <Text variant="subtext">{error?.error || 'Unknown error occurred'}</Text>
      </>
    ),
    onSuccess: () => {
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
      removeModal(props.modalId)

       if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: () => {
      trackEvent({
        event: 'action_run',
        user,
        status: 'error',
        props: {
          orgId: org?.id,
          installId: install?.id,
          actionConfigId: actionConfigId,
          err: error?.error,
        },
      })
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

    execute({
      body: {
        action_workflow_config_id: actionConfigId,
        ...(vars && Object.keys(vars)?.length > 0 && { run_env_vars: vars }),
      },
      installId: install?.id,
      orgId: org?.id,
      path,
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
      <Text>
        Are you sure you want to run the action {action?.name}?
      </Text>

      <form ref={formRef} onSubmit={handleSubmit} className="flex flex-col gap-4">
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
