import { useForm, useStore } from '@tanstack/react-form'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Divider } from '@/components/common/Divider'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Label } from '@/components/common/form/Label'
import { Toggle } from '@/components/common/form/Toggle'
import {
  installOperationRulesSchema,
  isRuleWindowValid,
  type InstallOperationRuleValues,
  type InstallOperationRulesValues,
  type TOperationRuleId,
  type TOperationWindowCadence,
  type TOutsideWindowPolicy,
} from './schema'

export interface IInstallOperationRules {
  installName: string
  timezone: string
  rules: InstallOperationRuleValues[]
  onCancel: () => void
  onSave: (values: InstallOperationRulesValues) => void
}

interface IOperationRuleMeta {
  label: string
  description: string
  icon: TIconVariant
}

const ruleMeta: Record<TOperationRuleId, IOperationRuleMeta> = {
  actions: {
    label: 'Actions',
    description:
      'One-off action runs triggered manually, on a schedule, or through the API.',
    icon: 'LightningIcon',
  },
  runbooks: {
    label: 'Runbooks',
    description: 'Ordered release procedures that combine deploys and actions.',
    icon: 'BookOpenTextIcon',
  },
  'sandbox-updates': {
    label: 'Sandbox updates',
    description:
      'Changes to the install sandbox — IAM, networking, and cluster infrastructure.',
    icon: 'FlaskIcon',
  },
  deployments: {
    label: 'Deployments',
    description: 'Component deploys and redeploys to this install.',
    icon: 'RocketIcon',
  },
  'break-glass': {
    label: 'Break glass',
    description: 'Emergency runs that bypass the normal operation path.',
    icon: 'LockKeyOpenIcon',
  },
}

const cadenceOptions: {
  value: TOperationWindowCadence
  label: string
}[] = [
  { value: 'anytime', label: 'Anytime' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
]

const weekdays = [
  { value: 'mon', label: 'Mon' },
  { value: 'tue', label: 'Tue' },
  { value: 'wed', label: 'Wed' },
  { value: 'thu', label: 'Thu' },
  { value: 'fri', label: 'Fri' },
  { value: 'sat', label: 'Sat' },
  { value: 'sun', label: 'Sun' },
]

const dayOfMonthOptions = [
  ...Array.from({ length: 28 }, (_, i) => ({
    value: `${i + 1}`,
    label: `Day ${i + 1}`,
  })),
  { value: 'last', label: 'Last day of the month' },
]

const timezoneOptions = [
  { value: 'UTC', label: 'UTC' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles' },
  { value: 'America/Denver', label: 'America/Denver' },
  { value: 'America/Chicago', label: 'America/Chicago' },
  { value: 'America/New_York', label: 'America/New_York' },
  { value: 'Europe/London', label: 'Europe/London' },
  { value: 'Europe/Berlin', label: 'Europe/Berlin' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo' },
  { value: 'Australia/Sydney', label: 'Australia/Sydney' },
]

const policyMeta: Record<
  TOutsideWindowPolicy,
  { label: string; description: string; icon: TIconVariant }
> = {
  reject: {
    label: 'Reject',
    description: 'Block the operation and notify whoever triggered it.',
    icon: 'ProhibitIcon',
  },
  approval: {
    label: 'Ask for permission',
    description: 'Hold the operation until an approver responds.',
    icon: 'ShieldCheckIcon',
  },
  queue: {
    label: 'Queue',
    description: 'Run automatically once the next window opens.',
    icon: 'ClockIcon',
  },
}

const policyOptions = (['reject', 'approval', 'queue'] as const).map(
  (value) => ({
    value,
    label: (
      <span className="flex flex-col gap-0.5">
        <Text flex weight="strong">
          <Icon variant={policyMeta[value].icon} size={14} />
          {policyMeta[value].label}
        </Text>
        <Text variant="subtext" theme="neutral">
          {policyMeta[value].description}
        </Text>
      </span>
    ),
  })
)

const windowSummary = (rule: InstallOperationRuleValues) => {
  if (rule.cadence === 'anytime') return 'anytime'
  const when =
    rule.cadence === 'weekly'
      ? rule.daysOfWeek.length
        ? rule.daysOfWeek
            .map(
              (day) => weekdays.find((weekday) => weekday.value === day)?.label
            )
            .join(', ')
        : 'no days selected'
      : rule.dayOfMonth === 'last'
        ? 'the last day of the month'
        : `day ${rule.dayOfMonth} of the month`
  return `${when}, ${rule.startTime}–${rule.endTime}`
}

const policyPhrases: Record<TOutsideWindowPolicy, [string, string]> = {
  reject: ['is rejected', 'are rejected'],
  approval: ['asks for permission', 'ask for permission'],
  queue: ['is queued', 'are queued'],
}

const policySummary = (counts: Record<TOutsideWindowPolicy, number>) => {
  const parts = (['reject', 'approval', 'queue'] as const)
    .filter((policy) => counts[policy])
    .map((policy) => {
      const count = counts[policy]
      const [singular, plural] = policyPhrases[policy]
      return `${count} ${count === 1 ? 'operation type' : 'operation types'} ${count === 1 ? singular : plural}`
    })
  if (parts.length < 2) return parts.join('')
  return `${parts.slice(0, -1).join(', ')} and ${parts.at(-1)}`
}

export const InstallOperationRules = ({
  installName,
  timezone,
  rules,
  onCancel,
  onSave,
}: IInstallOperationRules) => {
  const form = useForm({
    defaultValues: { timezone, rules } as InstallOperationRulesValues,
    validators: {
      onMount: installOperationRulesSchema,
      onChange: installOperationRulesSchema,
    },
    onSubmit: ({ value }) => onSave(value),
  })
  const values = useStore(form.store, (state) => state.values)
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  const enabledRules = values.rules.filter((rule) => rule.enabled)
  const allEnabled = values.rules.every((rule) => rule.enabled)
  const invalidRules = values.rules.filter((rule) => !isRuleWindowValid(rule))
  const restrictedRules = enabledRules.filter(
    (rule) => rule.cadence !== 'anytime' && rule.enforceOutsideWindow
  )
  const policyCounts = restrictedRules.reduce(
    (counts, rule) => ({
      ...counts,
      [rule.outsideWindowPolicy]: (counts[rule.outsideWindowPolicy] ?? 0) + 1,
    }),
    {} as Record<TOutsideWindowPolicy, number>
  )

  const setAllEnabled = (enabled: boolean) =>
    values.rules.forEach((_, index) =>
      form.setFieldValue(`rules[${index}].enabled`, enabled)
    )

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex items-start gap-3">
          <Button
            variant="ghost"
            size="sm"
            aria-label="Back to install"
            onClick={onCancel}
          >
            <Icon variant="ArrowLeftIcon" />
          </Button>
          <div className="flex flex-col items-start gap-1">
            <Text as="h1" variant="h2" weight="stronger">
              Install operation rules
            </Text>
            <Text theme="neutral">
              Control when operations may run on {installName}, and what happens
              when one is triggered outside its window.
            </Text>
            <Badge size="sm" theme="neutral">
              <Icon variant="CubeIcon" size={13} />
              {installName}
            </Badge>
          </div>
        </div>
        <Button
          variant="primary"
          disabled={!canSubmit || !!invalidRules.length}
          tooltipProps={
            invalidRules.length
              ? { tipContent: 'Finish the highlighted windows first' }
              : undefined
          }
          onClick={() => form.handleSubmit()}
        >
          <Icon variant="CheckIcon" />
          Save rules
        </Button>
      </div>

      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-6"
      >
        <Card className="!gap-4 !p-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-col gap-2">
              <Toggle
                checked={allEnabled}
                onChange={setAllEnabled}
                label="Enable all"
                description={`${enabledRules.length} of ${values.rules.length} operation types are governed by a rule.`}
              />
            </div>
            <form.Field name="timezone">
              {(field) => (
                <FormSelect
                  field={field}
                  size="sm"
                  options={timezoneOptions}
                  className="min-w-[16rem]"
                  labelProps={{ labelText: 'Window timezone' }}
                />
              )}
            </form.Field>
          </div>
        </Card>

        <Card className="!gap-4 !p-4">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              Operation rules
            </Text>
            <Text variant="subtext" theme="neutral">
              Give each operation type a repeating window, then choose what to
              do with operations triggered outside it.
            </Text>
          </div>

          <div className="-mx-4">
            <Divider />
          </div>

          <div className="divide-y">
            {values.rules.map((rule, index) => {
              const meta = ruleMeta[rule.id]
              const isAnytime = rule.cadence === 'anytime'
              const isValid = isRuleWindowValid(rule)

              return (
                <div
                  key={rule.id}
                  className="grid gap-4 py-5 first:pt-0 last:pb-0 sm:grid-cols-[minmax(0,1fr)_minmax(22rem,1fr)]"
                >
                  <div className="flex flex-col gap-2">
                    <Text flex weight="strong">
                      <Icon variant={meta.icon} size={16} />
                      {meta.label}
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      {meta.description}
                    </Text>
                    <form.Field name={`rules[${index}].enabled`}>
                      {(field) => (
                        <Toggle
                          checked={!!field.state.value}
                          onChange={(checked) => field.handleChange(checked)}
                          label={
                            field.state.value ? 'Rule enabled' : 'Not enforced'
                          }
                          aria-label={`${meta.label} rule`}
                        />
                      )}
                    </form.Field>
                  </div>

                  {!rule.enabled ? (
                    <Text variant="subtext" theme="neutral">
                      {meta.label} may run at any time.
                    </Text>
                  ) : (
                    <div className="flex flex-col gap-4">
                      <div className="flex flex-col gap-2">
                        <Label>Window</Label>
                        <form.Field name={`rules[${index}].cadence`}>
                          {(field) => (
                            <ToggleButton<TOperationWindowCadence>
                              options={cadenceOptions}
                              value={
                                field.state.value as TOperationWindowCadence
                              }
                              onChange={(value) => field.handleChange(value)}
                              className="self-start"
                            />
                          )}
                        </form.Field>
                      </div>

                      {isAnytime ? (
                        <Text flex variant="subtext" theme="neutral">
                          <Icon variant="CheckCircleIcon" size={14} />
                          {meta.label} may run at any time — nothing falls
                          outside this window.
                        </Text>
                      ) : (
                        <>
                          {rule.cadence === 'weekly' ? (
                            <form.Field name={`rules[${index}].daysOfWeek`}>
                              {(field) => {
                                const selected =
                                  (field.state.value as string[]) ?? []
                                return (
                                  <div className="flex flex-col gap-2">
                                    <Label>Days</Label>
                                    <div className="flex flex-wrap gap-1">
                                      {weekdays.map((day) => (
                                        <Button
                                          key={day.value}
                                          size="xs"
                                          variant={
                                            selected.includes(day.value)
                                              ? 'primary'
                                              : 'secondary'
                                          }
                                          aria-pressed={selected.includes(
                                            day.value
                                          )}
                                          className="min-w-[3rem] justify-center"
                                          onClick={() =>
                                            field.handleChange(
                                              selected.includes(day.value)
                                                ? selected.filter(
                                                    (value) =>
                                                      value !== day.value
                                                  )
                                                : [...selected, day.value]
                                            )
                                          }
                                        >
                                          {day.label}
                                        </Button>
                                      ))}
                                    </div>
                                  </div>
                                )
                              }}
                            </form.Field>
                          ) : (
                            <form.Field name={`rules[${index}].dayOfMonth`}>
                              {(field) => (
                                <FormSelect
                                  field={field}
                                  size="sm"
                                  options={dayOfMonthOptions}
                                  labelProps={{ labelText: 'Day' }}
                                />
                              )}
                            </form.Field>
                          )}

                          <div className="grid gap-3 sm:grid-cols-2">
                            <form.Field name={`rules[${index}].startTime`}>
                              {(field) => (
                                <FormInput
                                  field={field}
                                  type="time"
                                  size="sm"
                                  labelProps={{ labelText: 'Opens' }}
                                />
                              )}
                            </form.Field>
                            <form.Field name={`rules[${index}].endTime`}>
                              {(field) => (
                                <FormInput
                                  field={field}
                                  type="time"
                                  size="sm"
                                  labelProps={{ labelText: 'Closes' }}
                                />
                              )}
                            </form.Field>
                          </div>

                          <form.Field
                            name={`rules[${index}].enforceOutsideWindow`}
                          >
                            {(field) => (
                              <FormCheckbox
                                field={field}
                                labelProps={{
                                  className: '-ml-2',
                                  labelText:
                                    'When triggered outside of this window, do the following',
                                }}
                              />
                            )}
                          </form.Field>

                          {rule.enforceOutsideWindow ? (
                            <form.Field
                              name={`rules[${index}].outsideWindowPolicy`}
                            >
                              {(field) => (
                                <FormRadioGroup
                                  field={field}
                                  options={policyOptions}
                                />
                              )}
                            </form.Field>
                          ) : (
                            <Text variant="subtext" theme="neutral">
                              Operations triggered outside the window run
                              normally.
                            </Text>
                          )}

                          {!isValid ? (
                            <Text flex variant="subtext" theme="error">
                              <Icon variant="WarningIcon" size={14} />
                              {rule.cadence === 'weekly' &&
                              !rule.daysOfWeek.length
                                ? 'Pick at least one day.'
                                : 'Set an opening and closing time that differ.'}
                            </Text>
                          ) : null}
                        </>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      </form>

      <Card className="!gap-4 !p-4 bg-cool-grey-50 dark:bg-dark-grey-700">
        <div className="flex items-center gap-2">
          <Icon variant="InfoIcon" size={18} theme="brand" />
          <Text variant="base" weight="strong">
            Impact
          </Text>
        </div>
        <div className="flex flex-col gap-2">
          {enabledRules.length ? (
            <>
              {enabledRules.map((rule) => (
                <Text key={rule.id} flex variant="subtext">
                  <Icon variant={ruleMeta[rule.id].icon} size={15} />
                  {ruleMeta[rule.id].label}: {windowSummary(rule)}
                  {rule.cadence === 'anytime' ? '.' : ` (${values.timezone}).`}
                </Text>
              ))}
              <Text flex variant="subtext" theme="neutral">
                <Icon variant="ClockIcon" size={15} />
                {restrictedRules.length
                  ? `Outside their window, ${policySummary(policyCounts)}.`
                  : 'No operation type blocks work outside its window.'}
              </Text>
            </>
          ) : (
            <Text flex variant="subtext" theme="neutral">
              <Icon variant="CheckCircleIcon" size={15} />
              No rules are enforced — every operation may run at any time.
            </Text>
          )}
          <Text flex variant="subtext" theme="neutral">
            <Icon variant="ShieldIcon" size={15} />
            Break glass runs are permissioned and audited separately.
          </Text>
        </div>
      </Card>
    </div>
  )
}
