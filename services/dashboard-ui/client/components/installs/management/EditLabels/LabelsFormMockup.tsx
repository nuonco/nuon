import { useForm, useStore } from '@tanstack/react-form'
import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { editLabelsSchema, type EditLabelsValues } from './schema'

const INSTALL_LABELS_DOCS =
  'https://docs.nuon.co/guides/install-configs#labeling-installs'
const DEFAULT_LABELS_DOCS =
  'https://docs.nuon.co/guides/managing-apps#default-labels'

const COLLAPSED_ROW_COUNT = 5

interface ILabelsFormMockup extends Omit<IModal, 'onSubmit'> {
  labels: Record<string, string>
  defaultLabels?: Record<string, string>
  isPending?: boolean
  error?: TAPIError | null
  onSubmit: (labels: Record<string, string>) => void
}

export const LabelsFormMockup = ({
  labels: initialLabels,
  defaultLabels = {},
  isPending = false,
  error = null,
  onSubmit,
  ...props
}: ILabelsFormMockup) => {
  const [isExpanded, setIsExpanded] = useState(false)

  const form = useForm({
    defaultValues: {
      labels: Object.entries(initialLabels)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => ({ key, value })),
    } as EditLabelsValues,
    validators: { onMount: editLabelsSchema, onChange: editLabelsSchema },
    onSubmit: ({ value }) =>
      onSubmit(
        Object.fromEntries(
          value.labels.map((label) => [label.key.trim(), label.value.trim()])
        )
      ),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const defaultEntries = Object.entries(defaultLabels).sort(([a], [b]) =>
    a.localeCompare(b)
  )

  return (
    <Modal
      size="lg"
      childrenClassName="!pt-10 !gap-8"
      heading={
        <Text flex className="gap-3" variant="h3" weight="strong">
          <Icon variant="TagIcon" size="24" />
          Edit labels
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving labels
          </span>
        ) : (
          'Save labels'
        ),
        onClick: () => form.handleSubmit(),
        disabled: !canSubmit || isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-8"
      >
        <FormErrorBanner error={error} fallback="Unable to update labels" />

        <Text theme="neutral">
          Labels identify this install and target it from webhooks, Slack
          subscriptions, and deployment groups.{' '}
          <Link href={INSTALL_LABELS_DOCS} isExternal variant="inline">
            View label docs
          </Link>
        </Text>

        <form.Field name="labels" mode="array">
          {(labelsField) => {
            const rows = labelsField.state.value
            const hiddenCount = Math.max(rows.length - COLLAPSED_ROW_COUNT, 0)
            const isCollapsible = hiddenCount > 0

            return (
              <div className="flex flex-col gap-4">
                <div className="flex items-center justify-between gap-4">
                  <Text variant="base" weight="strong">
                    Install labels{rows.length > 0 ? ` (${rows.length})` : ''}
                  </Text>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    disabled={isPending}
                    onClick={() =>
                      labelsField.pushValue({ key: '', value: '' })
                    }
                  >
                    <Icon variant="PlusIcon" size="16" />
                    Add label
                  </Button>
                </div>

                {rows.length === 0 ? (
                  <Text variant="subtext" theme="neutral">
                    No labels yet. Add a label to identify this install.
                  </Text>
                ) : (
                  <div className="flex flex-col gap-3">
                    <div className="grid grid-cols-[1fr_1fr_auto] gap-3 items-center">
                      <Text variant="subtext" theme="neutral">
                        Key
                      </Text>
                      <Text variant="subtext" theme="neutral">
                        Value
                      </Text>
                      <span className="w-8" />
                    </div>

                    <div
                      id="mockup-install-labels"
                      className="flex flex-col gap-3"
                    >
                      {rows.map((_, idx) => (
                        <fieldset
                          key={idx}
                          className={
                            !isExpanded && idx >= COLLAPSED_ROW_COUNT
                              ? 'hidden'
                              : 'grid grid-cols-[1fr_1fr_auto] gap-3 items-start'
                          }
                        >
                          <form.Field name={`labels[${idx}].key`}>
                            {(field) => (
                              <FormInput
                                field={field}
                                type="text"
                                placeholder="env"
                                disabled={isPending}
                              />
                            )}
                          </form.Field>
                          <form.Field name={`labels[${idx}].value`}>
                            {(field) => (
                              <FormInput
                                field={field}
                                type="text"
                                placeholder="production"
                                disabled={isPending}
                              />
                            )}
                          </form.Field>
                          <Button
                            type="button"
                            variant="icon"
                            size="lg"
                            disabled={isPending}
                            aria-label={`Remove label ${idx + 1}`}
                            onClick={() => labelsField.removeValue(idx)}
                          >
                            <Icon variant="XIcon" size="16" />
                          </Button>
                        </fieldset>
                      ))}
                    </div>

                    {isCollapsible ? (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="w-full justify-center"
                        aria-expanded={isExpanded}
                        aria-controls="mockup-install-labels"
                        onClick={() => setIsExpanded((prev) => !prev)}
                      >
                        <Icon
                          variant={isExpanded ? 'CaretUpIcon' : 'CaretDownIcon'}
                          size="16"
                        />
                        {isExpanded
                          ? 'Show fewer labels'
                          : `Show ${hiddenCount} more labels`}
                      </Button>
                    ) : null}
                  </div>
                )}
              </div>
            )
          }}
        </form.Field>

        <Divider />

        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Default labels
            {defaultEntries.length > 0 ? ` (${defaultEntries.length})` : ''}
          </Text>

          {defaultEntries.length === 0 ? (
            <Text variant="subtext" theme="neutral">
              No default labels. Every install of this app inherits its default
              labels from the app config.{' '}
              <Link href={DEFAULT_LABELS_DOCS} isExternal variant="inline">
                View default label docs
              </Link>
            </Text>
          ) : (
            <>
              <div className="flex flex-wrap items-center gap-2">
                {defaultEntries.map(([key, value]) => (
                  <LabelBadge
                    key={`default:${key}`}
                    labelKey={key}
                    labelValue={value}
                  />
                ))}
              </div>
              <div className="flex items-start gap-1.5">
                <Icon
                  variant="LockIcon"
                  size="14"
                  className="mt-[2px] shrink-0 text-cool-grey-600 dark:text-white/70"
                />
                <Text variant="subtext" theme="neutral">
                  Owned by the app config and applied to every install.{' '}
                  <Link href={DEFAULT_LABELS_DOCS} isExternal variant="inline">
                    View default label docs
                  </Link>
                </Text>
              </div>
            </>
          )}
        </div>
      </form>
    </Modal>
  )
}
