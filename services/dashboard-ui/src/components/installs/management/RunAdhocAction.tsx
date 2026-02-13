'use client'

import { usePathname, useRouter } from 'next/navigation'
import { FormEvent, useEffect, useRef, useState } from 'react'
import { runAdhocAction } from '@/actions/installs/run-adhoc-action'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Editor } from '@/components/common/Editor'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TRunAdhocActionBody } from '@/lib'

interface IRunAdhocAction {}

export const RunAdhocActionModal = ({ ...props }: IRunAdhocAction & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal, addModal } = useSurfaces()

  const formRef = useRef<HTMLFormElement>(null)
  const draftShownRef = useRef(false)
  const [customEnvVars, setCustomEnvVars] = useState<number[]>([])
  const [inputMode, setInputMode] = useState<'command' | 'script'>('command')
  const [scriptContent, setScriptContent] = useState('')

  const {
    hasDraft,
    draftTimestamp,
    draftValues,
    clearDraft,
    restoreDraft,
    formKey,
  } = useFormPersistence({
    storageKey: `adhoc-action-draft:${install.id}`,
    formRef,
    enabled: true,
  })

  useEffect(() => {
    if (hasDraft && !draftShownRef.current && draftTimestamp) {
      draftShownRef.current = true

      let modalId: string
      const modal = (
        <ResumeDraftModal
          draftTimestamp={draftTimestamp}
          onResume={() => {
            restoreDraft()
            removeModal(modalId)
          }}
          onStartFresh={() => {
            clearDraft()
            draftShownRef.current = false
            removeModal(modalId)
          }}
          onClose={() => {
            removeModal(modalId)
          }}
        />
      )
      modalId = addModal(modal)
    }
  }, [hasDraft, draftTimestamp, restoreDraft, clearDraft, addModal, removeModal])

  useEffect(() => {
    if (draftValues) {
      if (draftValues['inputMode']) {
        setInputMode(draftValues['inputMode'] as 'command' | 'script')
      }
      if (draftValues['scriptContent']) {
        setScriptContent(draftValues['scriptContent'])
      }
      const customVarIndices = Object.keys(draftValues)
        .filter((key) => key.startsWith('custom:'))
        .map((key) => parseInt(key.split(':')[1]))
        .filter((val, idx, arr) => arr.indexOf(val) === idx)
      if (customVarIndices.length > 0) {
        setCustomEnvVars(customVarIndices)
      }
    }
  }, [draftValues])

  const { data, error, headers, isLoading, execute } = useServerAction({
    action: runAdhocAction,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unable to run adhoc action.</Text>,
    errorHeading: 'Adhoc action failed',
    onSuccess: () => {
      clearDraft()
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: <Text>Adhoc action is running.</Text>,
    successHeading: 'Adhoc action started',
  })

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const formDataObj = Object.fromEntries(formData)

    const env_vars = customEnvVars.reduce(
      (acc, cv) => {
        const name = formDataObj[`custom:${cv}:name`]
        const value = formDataObj[`custom:${cv}:value`]
        if (typeof name === 'string' && typeof value === 'string') {
          acc[name] = value
        }
        return acc
      },
      {} as Record<string, string>
    )

    const body: TRunAdhocActionBody = {
      name: (formDataObj.name as string) || undefined,
      timeout: formDataObj.timeout ? Number(formDataObj.timeout) : undefined,
      env_vars: Object.keys(env_vars).length > 0 ? env_vars : undefined,
    }

    if (inputMode === 'command') {
      body.command = formDataObj.command as string
    } else {
      body.inline_contents = scriptContent
    }

    execute({
      body,
      installId: install.id,
      orgId: org.id,
      path,
    })
  }

  return (
    <Modal
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="TerminalWindowIcon" size="24" />
          Run adhoc action
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Running action
          </span>
        ) : (
          'Run action'
        ),
        onClick: () => formRef.current?.requestSubmit(),
        disabled: isLoading,
        variant: 'primary',
      }}
      size="half"
      {...props}
    >
      <form
        key={formKey}
        ref={formRef}
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        {error && <Banner theme="error">{error?.error}</Banner>}

        <input type="hidden" name="inputMode" value={inputMode} />
        <input type="hidden" name="scriptContent" value={scriptContent} />

        <label className="flex flex-col gap-1">
          <Text variant="label" weight="strong">
            Name (optional)
          </Text>
          <Input
            name="name"
            type="text"
            placeholder="Display name for this action"
            maxLength={255}
            defaultValue={draftValues?.['name'] || ''}
          />
        </label>

        <div className="flex gap-2">
          <Button
            type="button"
            variant={inputMode === 'command' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setInputMode('command')}
          >
            Single Command
          </Button>
          <Button
            type="button"
            variant={inputMode === 'script' ? 'primary' : 'secondary'}
            size="sm"
            onClick={() => setInputMode('script')}
          >
            Bash Script
          </Button>
        </div>

        {inputMode === 'command' ? (
          <label className="flex flex-col gap-1">
            <Text variant="label" weight="strong">
              Command *
            </Text>
            <Input
              name="command"
              type="text"
              placeholder="echo 'Hello, world!'"
              required
              defaultValue={draftValues?.['command'] || ''}
            />
            <Text variant="subtext">Single-line shell command to execute</Text>
          </label>
        ) : (
          <label className="flex flex-col gap-1">
            <Text variant="label" weight="strong">
              Bash Script *
            </Text>
            <Editor
              value={scriptContent}
              onChange={setScriptContent}
              language="bash"
              placeholder="#!/bin/bash&#10;echo 'Hello, world!'"
              minHeight={200}
              maxHeight={600}
            />
            <Text variant="subtext">Multi-line bash script to execute</Text>
          </label>
        )}

        <label className="flex flex-col gap-1">
          <Text variant="label" weight="strong">
            Timeout (seconds)
          </Text>
          <Input
            name="timeout"
            type="number"
            defaultValue={draftValues?.['timeout'] || '300'}
            min={1}
            max={3600}
          />
          <Text variant="subtext">
            Execution timeout (1-3600 seconds, default: 300)
          </Text>
        </label>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="label" weight="strong">
              Environment Variables
            </Text>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                setCustomEnvVars((vars) => [...vars, vars.length])
              }}
            >
              <Icon variant="Plus" size="16" />
              Add variable
            </Button>
          </div>

          {customEnvVars.length === 0 && (
            <Text variant="subtext">No environment variables added</Text>
          )}

          {customEnvVars.map((cv) => (
            <fieldset
              key={cv}
              className="grid grid-cols-[1fr_1fr_auto] gap-2 items-end border-t pt-2"
            >
              <label className="flex flex-col gap-1">
                <Text variant="label">Name</Text>
                <Input
                  name={`custom:${cv}:name`}
                  type="text"
                  placeholder="VAR_NAME"
                  required
                  defaultValue={draftValues?.[`custom:${cv}:name`] || ''}
                />
              </label>
              <label className="flex flex-col gap-1">
                <Text variant="label">Value</Text>
                <Input
                  name={`custom:${cv}:value`}
                  type="text"
                  placeholder="value"
                  required
                  defaultValue={draftValues?.[`custom:${cv}:value`] || ''}
                />
              </label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => {
                  setCustomEnvVars((vars) => vars.filter((v) => v !== cv))
                }}
                className="mb-1"
              >
                <Icon variant="X" size="16" />
              </Button>
            </fieldset>
          ))}
        </div>
      </form>
    </Modal>
  )
}

export const RunAdhocActionButton = ({
  ...props
}: IRunAdhocAction & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RunAdhocActionModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Run adhoc action
      <Icon variant="TerminalWindowIcon" />
    </Button>
  )
}
