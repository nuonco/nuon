import { useState, useMemo, useEffect } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useOnboardingPoll } from '@/hooks/use-onboarding-poll'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Text } from '@/components/common/Text'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { cn } from '@/utils/classnames'
import { completeInstallStep, getApp } from '@/lib'
import type { TAPIError, TAppInputConfig, TOnboarding } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

function SelectionIndicator({ selected }: { selected: boolean }) {
  return (
    <div
      className={cn(
        'shrink-0 w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all',
        selected
          ? 'border-primary-600 bg-primary-600'
          : 'border-cool-grey-400 dark:border-cool-grey-600'
      )}
    >
      {selected && (
        <div className="w-2 h-2 rounded-full bg-white" />
      )}
    </div>
  )
}

const CARD_BASE =
  'flex flex-col w-full gap-3 p-5 rounded-md text-left transition-all cursor-pointer border bg-white dark:bg-dark-grey-900 shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]'
const CARD_DEFAULT =
  'border-cool-grey-500/24 dark:border-cool-grey-500/24'
const CARD_SELECTED =
  'border-primary-500 ring-2 ring-primary-500'

type CloudSetupOption = 'cloud' | 'sandbox'
type CloudPlatform = 'aws' | 'gcp' | 'azure'

const CLOUD_LABELS: Record<CloudPlatform, string> = {
  aws: 'AWS',
  gcp: 'GCP',
  azure: 'Azure',
}

function buildDefaultInputValues(inputConfig?: TAppInputConfig | null): Record<string, string> {
  if (!inputConfig?.input_groups) return {}
  const values: Record<string, string> = {}
  for (const group of inputConfig.input_groups) {
    for (const input of group.app_inputs ?? []) {
      if (input.name) {
        values[input.name] = input.default ?? ''
      }
    }
  }
  return values
}

type AppInputGroup = NonNullable<TAppInputConfig['input_groups']>[number]
type AppInput = NonNullable<AppInputGroup['app_inputs']>[number]

function InputField({
  input,
  value,
  onChange,
}: {
  input: AppInput
  value: string
  onChange: (name: string, value: string) => void
}) {
  const isBool =
    input?.type === 'bool' ||
    input?.default === 'true' ||
    input?.default === 'false'

  if (isBool) {
    return (
      <div className="flex items-center gap-3">
        <CheckboxInput
          checked={value === 'true'}
          onChange={(e) => onChange(input.name!, e.target.checked ? 'true' : 'false')}
          labelProps={{
            labelText: input.display_name || input.name || '',
            className: 'hover:!bg-transparent !px-0',
          }}
        />
        {input.description && (
          <Text variant="subtext" theme="neutral" className="flex-1">
            {input.description}
          </Text>
        )}
      </div>
    )
  }

  return (
    <Input
      labelProps={{ labelText: input.display_name || input.name || '' }}
      helperText={input.description}
      type={
        input.type === 'number'
          ? 'number'
          : input.sensitive
            ? 'password'
            : 'text'
      }
      autoComplete="off"
      required={input.required}
      value={value}
      onChange={(e) => onChange(input.name!, e.target.value)}
      placeholder={`Enter ${input.display_name?.toLowerCase() || 'value'}`}
    />
  )
}

function InputGroupSection({
  group,
  inputValues,
  onInputChange,
}: {
  group: NonNullable<TAppInputConfig['input_groups']>[0]
  inputValues: Record<string, string>
  onInputChange: (name: string, value: string) => void
}) {
  const allInputs = (group.app_inputs ?? []).sort(
    (a, b) => (a?.index ?? 0) - (b?.index ?? 0)
  )
  if (allInputs.length === 0) return null

  const requiredInputs = allInputs.filter((i) => i?.required)
  const optionalInputs = allInputs.filter((i) => !i?.required)

  return (
    <fieldset className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <Text weight="strong">
          {group.display_name || group.name}
        </Text>
        {group.description && (
          <Text variant="subtext" theme="neutral">
            {group.description}
          </Text>
        )}
      </div>

      {requiredInputs.map((input) => (
        <InputField
          key={input.id ?? input.name}
          input={input}
          value={inputValues[input.name!] ?? ''}
          onChange={onInputChange}
        />
      ))}

      {optionalInputs.length > 0 && (
        <Expand
          heading="Optional"
          headerClassName="!px-4 bg-code"
          id={`${group.id}-optional`}
          isOpen={requiredInputs.length === 0}
          className="border rounded-md"
        >
          <div className="flex flex-col gap-4 p-4 border-t">
            {optionalInputs.map((input) => (
              <InputField
                key={input.id ?? input.name}
                input={input}
                value={inputValues[input.name!] ?? ''}
                onChange={onInputChange}
              />
            ))}
          </div>
        </Expand>
      )}
    </fieldset>
  )
}

export const CloudSetupStepContainer = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [selected, setSelected] = useState<CloudSetupOption | null>(null)
  const [waiting, setWaiting] = useState(false)
  const [inputValues, setInputValues] = useState<Record<string, string>>({})
  const [inputsInitialized, setInputsInitialized] = useState(false)

  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id
  const appId = onboarding?.app_id
  const cloudPlatform = onboarding?.cloud_provider as CloudPlatform | null
  const cloudLabel = cloudPlatform ? CLOUD_LABELS[cloudPlatform] : null

  const { data: app } = useQuery({
    queryKey: ['app', appId],
    queryFn: () => getApp({ appId: appId!, orgId: orgId! }),
    enabled: !!appId && !!orgId,
    refetchInterval: (query) =>
      query.state.data?.input_config ? false : 3000,
  })

  const inputConfig = app?.input_config
  const sortedGroups = useMemo(
    () =>
      (inputConfig?.input_groups ?? [])
        .slice()
        .sort((a, b) => (a?.index ?? 0) - (b?.index ?? 0)),
    [inputConfig]
  )
  const hasInputs = sortedGroups.some(
    (g) => (g.app_inputs?.length ?? 0) > 0
  )

  useEffect(() => {
    if (inputConfig && !inputsInitialized) {
      setInputValues(buildDefaultInputValues(inputConfig))
      setInputsInitialized(true)
    }
  }, [inputConfig, inputsInitialized])

  const handleInputChange = (name: string, value: string) => {
    setInputValues((prev) => ({ ...prev, [name]: value }))
  }

  const { mutate: submit, isPending, error } = useMutation({
    mutationFn: () => {
      if (!orgId || !selected) throw new Error('Missing required data')

      const inputs =
        Object.keys(inputValues).length > 0
          ? inputValues
          : undefined

      return completeInstallStep({
        body: {
          name: onboarding?.example_app_slug
            ? `${onboarding.example_app_slug}-demo`
            : `${cloudPlatform ?? 'nuon'}-demo`,
          install_mode: selected,
          inputs,
        },
        orgId,
      })
    },
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'in-progress') {
        setWaiting(true)
      } else {
        onAdvance()
      }
    },
  })

  useOnboardingPoll({
    enabled: waiting,
    onResolved: (ob) => {
      setWaiting(false)
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'error') return
      onAdvance()
    },
  })

  const isWorking = isPending || waiting

  const requiredInputsMissing = useMemo(() => {
    if (!hasInputs) return false
    for (const group of sortedGroups) {
      for (const input of group.app_inputs ?? []) {
        if (input.required && input.source !== 'customer' && !inputValues[input.name!]?.trim()) {
          return true
        }
      }
    }
    return false
  }, [sortedGroups, hasInputs, inputValues])

  const handleAdvance = () => {
    if (!selected || isWorking || requiredInputsMissing) return
    submit()
  }

  return (
    <div className="flex flex-col gap-6">
      {error && (
        <Banner theme="error">
          {(error as TAPIError).error ?? 'Failed to create install'}
        </Banner>
      )}
      {onboarding?.status_v2?.status === 'error' && onboarding?.step_error && (
        <Banner theme="error">{onboarding.step_error}</Banner>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <button
          type="button"
          onClick={() => setSelected('cloud')}
          className={cn(
            CARD_BASE,
            selected === 'cloud' ? CARD_SELECTED : CARD_DEFAULT
          )}
        >
          <div className="flex items-center gap-4">
            {cloudLabel ? (
              <CloudPlatformDisplay
                platform={cloudPlatform!}
                colorVariant="color"
                displayVariant="icon-only"
                iconSize="36"
              />
            ) : (
              <Icon variant="CloudArrowUpIcon" size="24" />
            )}
            <Text variant="base" weight="strong" className="flex-1">
              Connect{' '}
              {cloudLabel ? `your ${cloudLabel} account` : 'a cloud account'}
            </Text>
            <SelectionIndicator selected={selected === 'cloud'} />
          </div>
          <Text variant="body" theme="neutral" className="whitespace-normal">
            {cloudLabel
              ? `Deploy directly to your ${cloudLabel} infrastructure.`
              : 'Deploy directly to your own cloud infrastructure.'}
          </Text>
        </button>

        <button
          type="button"
          onClick={() => setSelected('sandbox')}
          className={cn(
            CARD_BASE,
            selected === 'sandbox' ? CARD_SELECTED : CARD_DEFAULT
          )}
        >
          <div className="flex items-center gap-4">
            <Icon variant="FlaskIcon" size="24" />
            <Text variant="base" weight="strong" className="flex-1">
              Use demo mode
            </Text>
            <Badge size="sm" theme="info">
              Recommended
            </Badge>
            <SelectionIndicator selected={selected === 'sandbox'} />
          </div>
          <Text variant="body" theme="neutral" className="whitespace-normal">
            We'll spin up a managed demo environment — no cloud account
            needed.
          </Text>
        </button>
      </div>

      {selected && hasInputs && (
        <div className="flex flex-col gap-5">
          <Divider />
          <Text variant="h3" role="heading" level={3}>
            Configure your install
          </Text>
          {sortedGroups.map((group) => (
            <InputGroupSection
              key={group.id ?? group.name}
              group={group}
              inputValues={inputValues}
              onInputChange={handleInputChange}
            />
          ))}
        </div>
      )}

      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeftIcon" weight="bold" /> Back
          </Button>
        ) : (
          <div />
        )}
        <Button
          type="button"
          variant="primary"
          disabled={!selected || isWorking || requiredInputsMissing}
          onClick={handleAdvance}
        >
          {waiting ? 'Setting up install...' : isPending ? 'Creating...' : 'Continue'}{' '}
          {!isWorking && <Icon variant="CaretRightIcon" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
