'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, type FormEvent, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import {
  ArrowsClockwiseIcon,
  CheckIcon,
  PlusIcon,
  WarningOctagonIcon,
} from '@phosphor-icons/react'
import { runAction } from '@/actions/installs/run-action'
import { Button, type IButton } from '@/components/old/Button'
import { Expand } from '@/components/old/Expand'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Text } from '@/components/old/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import type { TActionConfig, TAction } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'

interface IActionTriggerButton extends Omit<IButton, 'className' | 'onClick'> {
  action: TAction
  actionConfigId: string
}

function normalizeEnvVars(steps: TActionConfig['steps']) {
  const envVars = steps.reduce((acc, step) => {
    const keys = Object.keys(step?.env_vars)
    if (keys?.length) {
      keys.forEach((key) => {
        if (!acc[key]) {
          acc[key] = step?.env_vars[key]
        }
      })
    }
    return acc
  }, {})

  return envVars
}

export const ActionTriggerButton: FC<IActionTriggerButton> = ({
  action,
  actionConfigId,
  ...props
}) => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const config = action?.configs?.[0]
  const envVars = normalizeEnvVars(config?.steps)
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  const [customVars, setCustomVars] = useState([])

  const {
    data: run,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({
    action: runAction,
  })

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'action_run',
        user,
        status: 'error',
        props: {
          orgId: org.id,
          installId: install.id,
          actionConfigId: actionConfigId,
          err: error?.error,
        },
      })
    }

    if (run) {
      trackEvent({
        event: 'action_run',
        user,
        status: 'ok',
        props: {
          orgId: org.id,
          installId: install.id,
          actionConfigId: actionConfigId,
        },
      })

      setIsKickedOff(true)
      setIsOpen(false)
    }
  }, [run, error, headers])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-3xl"
              heading={`Run action workflow ${action?.name}?`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4">
                {error ? (
                  <span className="flex items-center gap-3  w-full p-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-base font-medium">
                    <WarningOctagonIcon size="20" />{' '}
                    {error?.error || 'Unable to run action'}
                  </span>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to run the action workflow{' '}
                  {action?.name}?
                </Text>

                <form
                  className="flex flex-col gap-4"
                  onSubmit={(e: FormEvent<HTMLFormElement>) => {
                    e.preventDefault()
                    setIsKickedOff(true)

                    // Variables from the form. This will include overwritten values.
                    const overwrite = Object.fromEntries(
                      new FormData(e.currentTarget)
                    )

                    // Build list of run variables to send, to override default variable values.
                    const vars = Object.keys(overwrite).reduce((acc, key) => {
                      // If variable value is unchanged, do not add it to the overrides.
                      // We only want to include variables whose values have been changed.
                      if (overwrite[key] == envVars[key]) return acc

                      const customKey = key.split(':')
                      if (
                        customKey?.at(0) === 'custom' &&
                        customKey?.at(2) === 'name'
                      ) {
                        acc[overwrite[key] as string] =
                          overwrite[
                            `${customKey?.at(0)}:${customKey?.at(1)}:value`
                          ]
                      } else if (
                        customKey?.at(0) === 'custom' &&
                        customKey?.at(2) === 'value'
                      ) {
                        return acc
                      } else {
                        acc[key] = overwrite[key]
                      }

                      return acc
                    }, {})

                    execute({
                      body: {
                        action_workflow_config_id: actionConfigId,
                        ...(vars &&
                          Object.keys(vars)?.length > 0 && {
                            run_env_vars: vars,
                          }),
                      },
                      installId: install.id,
                      orgId: org.id,
                      path,
                    })
                  }}
                >
                  <Expand
                    id="action-env-vars"
                    heading={<Text variant="med-12">Edit env vars</Text>}
                    parentClass="border rounded"
                    headerClass="px-2 py-2"
                    expandContent={
                      <div className="p-4 border-t">
                        <Text>
                          Edit or add custom env vars for this manual action
                          workflow run.
                        </Text>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 my-4">
                          {Object.keys(envVars).map((envVar) => (
                            <label key={envVar} className="flex flex-col gap-1">
                              <Text variant="med-12">{envVar}</Text>
                              <input
                                className="px-3 py-2 text-base rounded border bg-black/5 dark:bg-white/5 shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300"
                                required
                                defaultValue={envVars[envVar]}
                                name={envVar}
                                type="text"
                              />
                            </label>
                          ))}
                        </div>
                        <div className="w-full">
                          {customVars.length
                            ? customVars.map((cv) => (
                                <fieldset
                                  key={cv}
                                  className="grid grid-cols-1 md:grid-cols-2 gap-2 py-2 border-t"
                                >
                                  <legend className="text-base font-medium pr-2">
                                    Custom env var {cv + 1}
                                  </legend>
                                  <label className="flex flex-col gap-1">
                                    <Text variant="med-12">Name</Text>
                                    <input
                                      className="px-3 py-2 text-base rounded border bg-black/5 dark:bg-white/5 shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300"
                                      required
                                      name={`custom:${cv + 1}:name`}
                                      type="text"
                                    />
                                  </label>
                                  <label className="flex flex-col gap-1">
                                    <Text variant="med-12">Value</Text>
                                    <input
                                      className="px-3 py-2 text-base rounded border bg-black/5 dark:bg-white/5 shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300"
                                      required
                                      name={`custom:${cv + 1}:value`}
                                      type="text"
                                    />
                                  </label>
                                </fieldset>
                              ))
                            : null}
                        </div>
                        <div>
                          <Button
                            className="text-sm gap-2 flex items-center"
                            onClick={() => {
                              setCustomVars((vars) => [...vars, vars.length])
                            }}
                            type="button"
                          >
                            <PlusIcon />
                            Add env var
                          </Button>
                        </div>
                      </div>
                    }
                  />
                  <div className="flex gap-3 justify-end">
                    <Button
                      onClick={() => {
                        setIsOpen(false)
                      }}
                      className="text-base"
                      type="reset"
                    >
                      Cancel
                    </Button>
                    <Button
                      className="text-base flex items-center gap-1"
                      variant="primary"
                      type="submit"
                      disabled={isLoading}
                    >
                      {isLoading ? (
                        <SpinnerSVG />
                      ) : isKickedOff ? (
                        <CheckIcon size="18" />
                      ) : (
                        <ArrowsClockwiseIcon size="18" />
                      )}{' '}
                      Run action workflow
                    </Button>
                  </div>
                </form>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !h-fit"
        onClick={() => {
          setIsOpen(true)
        }}
        {...props}
      >
        Run workflow
      </Button>
    </>
  )
}
