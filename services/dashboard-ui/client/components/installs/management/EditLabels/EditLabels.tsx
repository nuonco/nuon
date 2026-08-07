import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { editLabelsSchema, type EditLabelsValues } from './schema'

interface IEditLabelsModal extends Omit<IModal, 'onSubmit'> {
  labels: Record<string, string>
  defaultLabels?: Record<string, string>
  isPending: boolean
  error: TAPIError | null
  onSubmit: (labels: Record<string, string>) => void
}

export const EditLabelsModal = ({
  labels: initialLabels,
  defaultLabels = {},
  isPending,
  error,
  onSubmit,
  ...props
}: IEditLabelsModal) => {
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

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
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
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Unable to update labels" />

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="label" weight="strong">
              Labels
            </Text>
            <form.Field name="labels" mode="array">
              {(labelsField) => (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={isPending}
                  onClick={() => labelsField.pushValue({ key: '', value: '' })}
                >
                  <Icon variant="PlusIcon" size="16" />
                  Add label
                </Button>
              )}
            </form.Field>
          </div>

          <Text variant="subtext">
            Values can use the interpolation syntax, e.g.{' '}
            <code>{'{{ .nuon.cloud_account.aws.region }}'}</code>. Dynamic
            values update as install state changes.
          </Text>

          {Object.entries(defaultLabels)
            .sort(([a], [b]) => a.localeCompare(b))
            .map(([key, value]) => (
              <fieldset
                key={`default:${key}`}
                className="grid grid-cols-[1fr_1fr_auto] gap-2 items-end border-t pt-2 opacity-60"
              >
                <label className="flex flex-col gap-1">
                  <Text variant="label">Key</Text>
                  <Input name="" type="text" disabled defaultValue={key} />
                </label>
                <label className="flex flex-col gap-1">
                  <Text variant="label">Value</Text>
                  <Input name="" type="text" disabled defaultValue={value} />
                </label>
                <Tooltip
                  tipContent="Defined in the app config's default_labels"
                  position="left"
                  tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs"
                >
                  <span className="mb-3 flex">
                    <Icon variant="LockIcon" size="16" />
                  </span>
                </Tooltip>
              </fieldset>
            ))}

          <form.Field name="labels" mode="array">
            {(labelsField) =>
              labelsField.state.value.length === 0 ? (
                Object.keys(defaultLabels).length === 0 ? (
                  <Text variant="subtext">No labels added</Text>
                ) : null
              ) : (
                <>
                  {labelsField.state.value.map((_, idx) => (
                    <fieldset
                      key={idx}
                      className="grid grid-cols-[1fr_1fr_auto] gap-2 items-end border-t pt-2"
                    >
                      <form.Field name={`labels[${idx}].key`}>
                        {(field) => (
                          <FormInput
                            field={field}
                            type="text"
                            placeholder="e.g. env"
                            disabled={isPending}
                            labelProps={{ labelText: 'Key' }}
                          />
                        )}
                      </form.Field>
                      <form.Field name={`labels[${idx}].value`}>
                        {(field) => (
                          <FormInput
                            field={field}
                            type="text"
                            placeholder="e.g. production"
                            disabled={isPending}
                            labelProps={{ labelText: 'Value' }}
                          />
                        )}
                      </form.Field>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        disabled={isPending}
                        onClick={() => labelsField.removeValue(idx)}
                        className="mb-1"
                      >
                        <Icon variant="XIcon" size="16" />
                      </Button>
                    </fieldset>
                  ))}
                </>
              )
            }
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
