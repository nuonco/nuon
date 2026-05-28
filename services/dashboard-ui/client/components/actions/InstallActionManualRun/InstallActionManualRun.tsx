import React, { useRef, useState, type FormEvent, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAction, TActionConfig } from '@/types'

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

interface IInstallActionManualRunModal extends Omit<IModal, 'onSubmit'> {
  action: TAction
  actionConfigId: string
  isLoading: boolean
  onSubmit: (vars: Record<string, string>, role: string) => void
  roleSelector: ReactNode
}

export const InstallActionManualRunModal = ({
  action,
  actionConfigId,
  isLoading,
  onSubmit,
  roleSelector,
  ...props
}: IInstallActionManualRunModal) => {
  const config = action?.configs?.[0]
  const envVars = normalizeEnvVars(config?.steps || [])
  const hasEnvVars = Object.keys(envVars).length > 0

  const [customVars, setCustomVars] = useState<number[]>([])
  const formRef = useRef<HTMLFormElement>(null)

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

    const roleInput = e.currentTarget.querySelector<HTMLInputElement>('[name="role"]')
    onSubmit(vars, roleInput?.value || '')
  }

  const handlePrimaryAction = () => {
    formRef.current?.requestSubmit()
  }

  return (
    <Modal
      heading={`Run action ${action?.name}?`}
      size="lg"
      primaryActionTrigger={{
        children: isLoading ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Running action...
          </>
        ) : (
          <>
            <Icon variant="PlayIcon" />
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
        {roleSelector}
        <Expand
          id="action-env-vars"
          heading={<Text weight="strong">Edit environment variables</Text>}
          className="border rounded-md"
          isOpen={hasEnvVars}
        >
          <div className="p-4 border-t flex flex-col gap-4">
            <Text variant="subtext">
              Edit or add custom environment variables for this manual action
              workflow run.
            </Text>

            {Object.keys(envVars).length > 0 && (
              <div className="flex flex-col gap-4">
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
                    className="flex flex-col gap-2 py-2 border-t relative"
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
                        <Icon variant="XIcon" size="12" />
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
                <Icon variant="PlusIcon" />
                Add environment variable
              </Button>
            </div>
          </div>
        </Expand>

      </form>
    </Modal>
  )
}
