'use client'

import React, { type FC, type FormEvent, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import {
  ArrowsClockwise,
  Check,
  Plus,
  WarningOctagon,
} from '@phosphor-icons/react'
import { Button, type IButton } from '@/components/Button'
import { Expand } from '@/components/Expand'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { runManualWorkflow } from './workflow-actions'
import type { TActionConfig, TActionWorkflow } from '@/types'
import { trackEvent } from '@/utils'

interface IActionTriggerButton extends Omit<IButton, 'className' | 'onClick'> {
  installId: string
  orgId: string
  workflowConfigId: string
  actionWorkflow: TActionWorkflow
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
  actionWorkflow,
  installId,
  orgId,
  workflowConfigId,
  ...props
}) => {
  const { user } = useUser()
  const config = actionWorkflow?.configs?.[0]
  const envVars = normalizeEnvVars(config?.steps)
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [customVars, setCustomVars] = useState([])
  const [error, setError] = useState()

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-3xl"
              heading={`Run action workflow ${actionWorkflow?.name}?`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4">
                {error ? (
                  <span className="flex items-center gap-3  w-full p-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-base font-medium">
                    <WarningOctagon size="20" /> {error}
                  </span>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to run the action workflow{' '}
                  {actionWorkflow?.name}?
                </Text>

                <form
                  className="flex flex-col gap-4"
                  onSubmit={(e: FormEvent<HTMLFormElement>) => {
                    e.preventDefault()
                    setIsLoading(true)
                    const overwrite = Object.fromEntries(
                      new FormData(e.currentTarget)
                    )
                    const vars = Object.keys(overwrite).reduce((acc, key) => {
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

                    runManualWorkflow({
                      installId,
                      orgId,
                      workflowConfigId,
                      vars,
                    })
                      .then(() => {
                        trackEvent({
                          event: 'action_run',
                          user,
                          status: 'ok',
                          props: { orgId, installId, workflowConfigId, vars },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      })
                      .catch((err) => {
                        trackEvent({
                          event: 'action_run',
                          user,
                          status: 'error',
                          props: {
                            orgId,
                            installId,
                            workflowConfigId,
                            vars,
                            err,
                          },
                        })
                        setError(
                          err?.message ||
                            'Error occured, please refresh page and try again.'
                        )
                        setIsLoading(false)
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
                            <Plus />
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
                      {isKickedOff ? (
                        <Check size="18" />
                      ) : isLoading ? (
                        <SpinnerSVG />
                      ) : (
                        <ArrowsClockwise size="18" />
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
