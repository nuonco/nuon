'use client'

import classNames from 'classnames'
import React, { type FC, type FormEvent, useRef, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckCircleIcon, CubeIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { CodeViewer } from '@/components/old/Code'
import { Link } from '@/components/old/Link'
import { Empty } from '@/components/old/Empty'
import { CheckboxInput, Input, RadioInput } from '@/components/old/Input'
import { SpinnerSVG, Loading } from '@/components/old/Loading'
import { Notice } from '@/components/old/Notice'
import { Select } from '@/components/old/Select'
import { Text } from '@/components/old/Typography'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { useOrg } from '@/hooks/use-org'
import { trackEvent } from '@/lib/segment-analytics'
import { getFlagEmoji } from '@/utils'
import type { TAppInputConfig, TInstall, TAPIResponse } from '@/types'

interface IInstallForm {
  platform?: string | 'aws' | 'azure'
  inputConfig?: TAppInputConfig
  install?: TInstall
  onSubmit: (
    formData: FormData
  ) => Promise<
    | TInstall
    | string
    | Record<'installId' | 'workflowId', string>
    | { error: any }
  >
  onSuccess: (res: TAPIResponse<TInstall>) => void
  onCancel: () => void
}

export const InstallForm: FC<IInstallForm> = ({
  inputConfig,
  install,
  platform,
  ...props
}) => {
  const { user } = useUser()
  const { org } = useOrg()
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
          'flex-auto flex flex-col gap-8 justify-between focus:outline-none relative pt-6'
        )}
        onKeyDown={handleTabChange}
        ref={formRef}
        onSubmit={(e: FormEvent<HTMLFormElement>) => {
          e.preventDefault()
          setIsLoading(true)
          const formData = new FormData(e.currentTarget)

          props
            .onSubmit(formData)
            .then((ins) => {
              if ((ins as any)?.error) {
                const errMsg =
                  (ins as any)?.error?.error ===
                  'unable to create install: unable to create install: duplicated key not allowed'
                    ? "Can't create install: duplicate install names not allowed"
                    : (ins as any)?.error?.error
                trackEvent({
                  event: install ? 'install_update' : 'install_create',
                  user,
                  status: 'error',
                  props: {
                    orgId: org.id,
                    installId: install ? install?.id : null,
                  },
                })
                setIsLoading(false)
                setError(
                  errMsg ||
                    'Unable to create install, refresh the page and try again.'
                )
                formRef.current?.parentElement?.scrollTo({
                  top: 0,
                  behavior: 'smooth',
                })
              } else {
                trackEvent({
                  event: install ? 'install_update' : 'install_create',
                  user,
                  status: 'ok',
                  props: {
                    orgId: org.id,
                    installId: install ? install?.id : (ins as TInstall)?.id,
                  },
                })
                setIsLoading(false)
                setIsCreated(true)
                props.onSuccess(ins as any)
              }
            })
            .catch((err) => {
              trackEvent({
                event: install ? 'install_update' : 'install_create',
                user,
                status: 'error',
                props: {
                  orgId: org.id,
                  installId: install ? install?.id : null,
                },
              })
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
            <Notice> {error}</Notice>
          </div>
        ) : null}
        {isLoading || isCreated ? (
          <div className="flex flex-auto items-center justify-center absolute w-full bg-black/10 dark:bg-black/70 h-full z-30 top-0 on-enter">
            {isLoading ? (
              <Loading loadingText="Creating install..." variant="page" />
            ) : null}
            {isCreated ? (
              <div className="flex flex-col gap-4 items-center on-enter">
                <CheckCircleIcon
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
          {install ? null : (
            <Field labelText="Install name">
              <Input
                type="text"
                name="name"
                defaultValue={install?.name}
                required
              />
            </Field>
          )}
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

        <div className="flex gap-3 justify-end border-t w-full p-6">
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
            {isLoading ? <SpinnerSVG /> : <CubeIcon size="16" />}{' '}
            {install ? 'Update' : 'Create'} Install
          </Button>
        </div>
      </form>
    </>
  )
}

const AWSFields: FC<{ cfLink?: string }> = ({ cfLink }) => {
  const options = AWS_REGIONS.map((o) => ({
    value: o.value,
    label: o?.iconVariant
      ? `${getFlagEmoji(o.iconVariant.substring(5))} ${o.text} [${o.value}]`
      : o.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t">
      <legend className="text-lg font-semibold mb-6 pr-6">
        Set AWS settings
      </legend>

      {/* <div className="max-w-lg mb-6">
          <Text variant="med-14">Create IAM policies with CloudFormation</Text>
          <Text className="!leading-relaxed !inline-block my-3">
          You can create a 1-click IAM role with the correct policies attached
          to provision + deprovision your application using the following link.
          This will create an IAM role granting access to install . Please use
          the stack output called{' '}
          <Code
          className="!py-1 !px-1.5 !bg-black/10 dark:!bg-white/10 leading-none !text-[10px] !inline-block align-middle"
          variant="inline"
          >
          RoleARN
          </Code>{' '}
          in the AWS IAM role input below.
          </Text>
          <Link
          className="text-sm"
          href={cfLink}
          target="_blank"
          onClick={() => {
          trackEvent({
          event: 'iam_role_create',
          user,
          status: 'ok',
          props: {
          orgId: org?.id,
          cfLink,
          },
          })
          }}
          >
          Create IAM Role
          </Link>
          </div>

          <Field labelText="Provide a resouce name for AWS IAM role *">
          <Input type="text" name="iam_role_arn" required />
          </Field> */}

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
    </fieldset>
  )
}

const InputConfigs: FC<{
  inputConfig: TAppInputConfig
  install?: TInstall
}> = ({ inputConfig, install }) => {
  return (
    <>
      {inputConfig?.input_groups ? (
        inputConfig?.input_groups
          ?.sort((a, b) => a?.index - b.index)
          ?.map((group) => (
            <InputGroupFields
              key={group.id}
              groupInputs={group}
              install={install}
            />
          ))
      ) : install ? (
        <div className="flex flex-col items-center ml-auto mr-40">
          <Empty
            variant="diagram"
            emptyTitle="Install has no inputs"
            emptyMessage="Add inputs to your app config."
          />
          <Link href="https://docs.nuon.co/concepts/app-inputs" isExternal>
            Learn more
          </Link>
        </div>
      ) : null}
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
      <legend className="flex flex-col gap-0 mb-6 pr-6">
        <span className="text-lg font-semibold">
          {groupInputs?.display_name}
        </span>
        <span className="text-sm font-normal">{groupInputs?.description}</span>
      </legend>

      {groupInputs?.app_inputs
        ?.sort((a, b) => a?.index - b?.index)
        ?.map((input) =>
          Boolean(input?.default === 'true' || input?.default === 'false') ||
          input?.type === 'bool' ? (
            <div
              key={input?.id}
              className="grid grid-cols-1 md:grid-cols-2 gap-4 items-start"
            >
              <div />
              <div className="ml-1">
                <input
                  type="hidden"
                  name={`inputs:${input?.name}`}
                  value="off"
                />
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
            </div>
          ) : (
            <Field
              key={input?.id}
              labelText={`${input?.display_name}${input?.required ? ' *' : ''}`}
              helpText={input?.description}
            >
              {input?.type === 'json' ? (
                <CodeViewer
                  language="json"
                  initCodeSource={
                    installInputs?.[input?.name] || input?.default
                  }
                  isEditable
                  name={`inputs:${input?.name}`}
                  required={input?.required}
                />
              ) : (
                <Input
                  type={
                    input?.type === 'number'
                      ? 'number'
                      : input?.sensitive
                        ? 'password'
                        : 'text'
                  }
                  autoComplete="off"
                  name={`inputs:${input?.name}`}
                  required={input?.required}
                  defaultValue={installInputs?.[input?.name] || input?.default}
                />
              )}
            </Field>
          )
        )}
    </fieldset>
  )
}

const UpdateInstallOptions: FC = () => {
  return (
    <div className="flex flex-col gap-0 px-6 max-w-3xl">
      <fieldset className="flex flex-col gap-4 border-t">
        <legend className="flex flex-col gap-0 mb-4 pr-6">
          <span className="text-lg font-semibold">Update install resouces</span>
          <span className="text-sm font-normal">
            Reprovision sandbox and redeploy components after updating install
            settings
          </span>
        </legend>

        <div className="flex gap-6 justify-start">
          <RadioInput
            name="form-control:update"
            labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit"
            labelText="Skip updating resources"
            defaultChecked
            value="skip"
          />
          <RadioInput
            name="form-control:update"
            labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit"
            labelText="Update all resources"
            value="update"
          />
        </div>
      </fieldset>
    </div>
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
