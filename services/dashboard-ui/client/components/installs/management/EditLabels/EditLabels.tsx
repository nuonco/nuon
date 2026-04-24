import { type FormEvent, useRef, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IEditLabelsModal extends Omit<IModal, 'onSubmit'> {
  labels: Record<string, string>
  isPending: boolean
  error: any
  onSubmit: (labels: Record<string, string>) => void
}

export const EditLabelsModal = ({
  labels: initialLabels,
  isPending,
  error,
  onSubmit,
  ...props
}: IEditLabelsModal) => {
  const formRef = useRef<HTMLFormElement>(null)

  const initialEntries = Object.entries(initialLabels).sort(([a], [b]) =>
    a.localeCompare(b),
  )
  const [rows, setRows] = useState<number[]>(initialEntries.map((_, i) => i))
  const nextId = useRef(initialEntries.length)

  const initialValues: Record<string, string> = {}
  initialEntries.forEach(([key, value], i) => {
    initialValues[`label:${i}:key`] = key
    initialValues[`label:${i}:value`] = value
  })

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const formDataObj = Object.fromEntries(formData)

    const labels = rows.reduce(
      (acc, idx) => {
        const key = (formDataObj[`label:${idx}:key`] as string)?.trim()
        const value = (formDataObj[`label:${idx}:value`] as string)?.trim()
        if (key) {
          acc[key] = value || ''
        }
        return acc
      },
      {} as Record<string, string>,
    )

    onSubmit(labels)
  }

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="info">
          <Icon variant="TagIcon" size="24" />
          Edit labels
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save labels'
        ),
        onClick: () => formRef.current?.requestSubmit(),
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <form
        ref={formRef}
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update labels'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="label" weight="strong">
              Labels
            </Text>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                setRows((r) => [...r, nextId.current++])
              }}
            >
              <Icon variant="Plus" size="16" />
              Add label
            </Button>
          </div>

          {rows.length === 0 && (
            <Text variant="subtext">No labels added</Text>
          )}

          {rows.map((idx) => (
            <fieldset
              key={idx}
              className="grid grid-cols-[1fr_1fr_auto] gap-2 items-end border-t pt-2"
            >
              <label className="flex flex-col gap-1">
                <Text variant="label">Key</Text>
                <Input
                  name={`label:${idx}:key`}
                  type="text"
                  placeholder="e.g. env"
                  required
                  defaultValue={initialValues[`label:${idx}:key`] || ''}
                />
              </label>
              <label className="flex flex-col gap-1">
                <Text variant="label">Value</Text>
                <Input
                  name={`label:${idx}:value`}
                  type="text"
                  placeholder="e.g. production"
                  defaultValue={initialValues[`label:${idx}:value`] || ''}
                />
              </label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => {
                  setRows((r) => r.filter((v) => v !== idx))
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
