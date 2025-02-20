'use client'

import classNames from 'classnames'
import React, { type FC, type FormEvent, useRef, useState } from 'react'
import { WarningOctagon, CheckCircle, Cube } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Input, CheckboxInput } from '@/components/Input'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Select } from '@/components/Select'
import { Text } from '@/components/Typography'
import { getFlagEmoji, AWS_REGIONS, AZURE_REGIONS } from '@/utils'
import type { TAppInputConfig, TInstall } from '@/types'

interface IInstallForm {
  platform?: string | 'aws' | 'azure'
  inputConfig?: TAppInputConfig
  install?: TInstall
  onSubmit: (formData: FormData) => Promise<TInstall>
  onSuccess: (install: TInstall) => void
  onCancel: () => void
}

export const InstallForm: FC<IInstallForm> = ({
  inputConfig,
  install,
  platform,
  ...props
}) => {
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isCreated, setIsCreated] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const handleTabChange = (event: any) => {
    if (event.key !== 'Tab') return

    event.preventDefault()

    // Get all focusable elements within the modal
    const focusableElements: any = formRef.current?.querySelectorAll(
      'button, select, textarea, [tabindex]:not([tabindex="-1"]):not(:disabled), input:not([type="hidden"])'
    )

    const firstFocusableElement = focusableElements?.[0]
    const lastFocusableElement =
      focusableElements?.[focusableElements.length - 1]

    // If the shift key is pressed and the first element is focused, move focus to the last element
    if (event.shiftKey && document.activeElement === firstFocusableElement) {
      lastFocusableElement?.focus()
      return
    }

    // If the shift key is not pressed and the last element is focused, move focus to the first element
    if (!event.shiftKey && document.activeElement === lastFocusableElement) {
      firstFocusableElement?.focus()
      return
    }

    // Otherwise, move focus to the next element
    const direction = event.shiftKey ? -1 : 1
    const index = Array.prototype.indexOf.call(
      focusableElements,
      document.activeElement
    )
    const nextElement = focusableElements?.[index + direction]
    if (nextElement) {
      nextElement?.focus()
    }
  }

  return (
    <>
      <form
        className={classNames(
          'min-h-[600px] flex-auto flex flex-col gap-8 justify-between focus:outline-none relative pt-6'
        )}
        onKeyDown={handleTabChange}
        ref={formRef}
        onSubmit={(e: FormEvent<HTMLFormElement>) => {
          e.preventDefault()
          setIsLoading(true)
          const formData = new FormData(e.currentTarget)

          props
            .onSubmit(formData)
            .then((install) => {
              setIsLoading(false)
              setIsCreated(true)
              props.onSuccess(install)
            })
            .catch(() => {
              setIsLoading(false)
              setError(
                'Unable to create install, refresh the page and try again.'
              )
              formRef.current?.parentElement?.scrollTo({
                top: 0,
                behavior: 'smooth',
              })
            })
        }}
      >
        {error ? (
          <div className="px-6">
            <span className="flex items-center gap-3 w-full p-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-base font-medium">
              <WarningOctagon size={50} /> {error}
            </span>
          </div>
        ) : null}
        {isLoading || isCreated ? (
          <div className="flex flex-auto items-center justify-center absolute w-full bg-black/10 dark:bg-black/70 h-full z-30 top-0 on-enter">
            {isLoading ? (
              <Loading loadingText="Creating install..." variant="page" />
            ) : null}
            {isCreated ? (
              <div className="flex flex-col gap-4 items-center on-enter">
                <CheckCircle
                  className="text-green-800 dark:text-green-500"
                  size={32}
                />
                <Text variant="reg-14">Install created, redirecting...</Text>
              </div>
            ) : null}
          </div>
        ) : null}
        <div
          className={classNames('flex flex-col gap-8 px-6 max-w-3xl', {
            blur: isLoading || isCreated,
          })}
        >
          <Field labelText="Install name">
            <Input
              type="text"
              name="name"
              defaultValue={install?.name}
              required
            />
          </Field>
          {platform ? (
            platform === 'aws' ? (
              <AWSFields />
            ) : (
              <AzureFields />
            )
          ) : null}
          {inputConfig ? (
            <InputConfigs inputConfig={inputConfig} install={install} />
          ) : null}
        </div>

        <div className="flex gap-3 justify-end border-t w-full p-6 p-6">
          <Button
            className="text-sm font-medium"
            type="reset"
            onClick={() => {
              setError(null)
              setIsLoading(false)
              setIsCreated(false)
              formRef.current?.reset()
              props.onCancel()
            }}
          >
            Cancel
          </Button>
          <Button
            className="flex items-center gap-1 text-sm font-medium disabled:!bg-primary-950"
            type="submit"
            variant="primary"
            disabled={isLoading}
          >
            {isLoading ? <SpinnerSVG /> : <Cube size="16" />}{' '}
            {install ? 'Update' : 'Create'} Install
          </Button>
        </div>
      </form>
    </>
  )
}

const AWSFields: FC = ({}) => {
  const options = AWS_REGIONS.map((o) => ({
    value: o.value,
    label: o?.iconVariant
      ? `${getFlagEmoji(o.iconVariant.substring(5))} ${o.text}`
      : o.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t">
      <legend className="text-lg font-semibold mb-6 pr-6">
        Set AWS settings
      </legend>

      <Field labelText="Provide a resouce name for AWS IAM role *">
        <Input type="text" name="iam_role_arn" required />
      </Field>

      <Field labelText="Select AWS region *">
        <Select name="region" options={options} required />
      </Field>
    </fieldset>
  )
}

const AzureFields: FC = ({}) => {
  const options = AZURE_REGIONS.map((o) => ({
    value: o.value,
    label: o?.iconVariant
      ? `${getFlagEmoji(o.iconVariant.substring(5))} ${o.text}`
      : o.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t">
      <legend className="text-lg font-semibold mb-6 pr-6">
        Set Azure configuration
      </legend>

      <Field labelText="Select Azure location *">
        <Select name="location" options={options} required />
      </Field>

      <Field labelText="Provide a service principal app ID *">
        <Input type="text" name="service_principal_app_id" required />
      </Field>

      <Field labelText="Provide a service principal password *">
        <Input type="text" name="service_principal_password" required />
      </Field>

      <Field labelText="Provide a subscription ID *">
        <Input type="text" name="subscription_id" required />
      </Field>

      <Field labelText="Provide a subscription tenant ID *">
        <Input type="text" name="subscription_tenant_id" required />
      </Field>
    </fieldset>
  )
}

const InputConfigs: FC<{
  inputConfig: TAppInputConfig
  install?: TInstall
}> = ({ inputConfig, install }) => {
  return (
    <>
      {inputConfig?.input_groups
        ? inputConfig?.input_groups?.map((group, i) => (
            <InputGroupFields
              key={group.id}
              groupInputs={group}
              install={install}
            />
          ))
        : null}
    </>
  )
}

const InputGroupFields: FC<{
  groupInputs: TAppInputConfig['input_groups'][0]
  install?: TInstall
}> = ({ groupInputs, install }) => {
  const installInputs = install ? install?.install_inputs?.at(0)?.values : {}

  return (
    <fieldset className="flex flex-col gap-6 border-t">
      <legend className="flex flex-col gap-0  mb-6 pr-6">
        <span className="text-lg font-semibold">
          {groupInputs?.display_name}
        </span>
        <span className="text-sm font-normal">{groupInputs?.description}</span>
      </legend>

      {groupInputs?.app_inputs?.map((input) =>
        Boolean(input?.default === 'true' || input?.default === 'false') ? (
          <div
            key={input?.id}
            className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start"
          >
            <div />
            <CheckboxInput
              labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0"
              labelTextClassName="!text-base !font-normal"
              defaultChecked={
                installInputs?.[input?.name]
                  ? Boolean(installInputs?.[input?.name] === 'true')
                  : Boolean(input?.default === 'true')
              }
              labelText={input?.display_name}
              name={`inputs:${input?.name}`}
            />
          </div>
        ) : (
          <Field
            key={input?.id}
            labelText={`${input?.display_name}${input?.required ? ' *' : ''}`}
            helpText={input?.description}
          >
            <Input
              type={input?.sensitive ? 'password' : 'text'}
              name={`inputs:${input?.name}`}
              required={input?.required}
              defaultValue={installInputs?.[input?.name] || input?.default}
            />
          </Field>
        )
      )}
    </fieldset>
  )
}

const Field: FC<{
  children: React.ReactElement
  labelText: string
  helpText?: string
}> = ({ children, labelText, helpText }) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="med-14">{labelText}</Text>
        {helpText ? (
          <Text variant="reg-12" className="max-w-72">
            {helpText}
          </Text>
        ) : null}
      </span>
      {children}
    </label>
  )
}
