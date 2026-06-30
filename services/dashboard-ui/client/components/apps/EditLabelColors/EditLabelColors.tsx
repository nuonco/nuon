import { type FormEvent, useRef, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IEditLabelColorsModal extends Omit<IModal, 'onSubmit'> {
  labelColors: Record<string, string>
  defaultColors?: string[]
  isPending: boolean
  error: any
  onSubmit: (labelColors: Record<string, string>) => void
}

const SWATCH_COLORS = [
  '#2563eb', '#dc2626', '#16a34a', '#9333ea', '#ca8a04', '#0891b2',
  '#e11d48', '#4f46e5', '#059669', '#c026d3', '#d97706', '#0284c7',
  '#7c3aed', '#15803d', '#a21caf', '#b45309', '#6366f1', '#ef4444',
  '#22c55e', '#a855f7', '#eab308', '#06b6d4', '#f43f5e', '#818cf8',
]

export const EditLabelColorsModal = ({
  labelColors: initialLabelColors,
  defaultColors,
  isPending,
  error,
  onSubmit,
  ...props
}: IEditLabelColorsModal) => {
  const formRef = useRef<HTMLFormElement>(null)

  const initialEntries = Object.entries(initialLabelColors).sort(([a], [b]) =>
    a.localeCompare(b),
  )
  const [rows, setRows] = useState<number[]>(initialEntries.map((_, i) => i))
  const [rowColors, setRowColors] = useState<Record<number, string>>(() => {
    const m: Record<number, string> = {}
    initialEntries.forEach(([, color], i) => {
      m[i] = color
    })
    return m
  })
  const nextId = useRef(initialEntries.length)

  const initialValues: Record<string, string> = {}
  initialEntries.forEach(([key, color], i) => {
    initialValues[`lc:${i}:key`] = key
    initialValues[`lc:${i}:color`] = color
  })

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const formDataObj = Object.fromEntries(formData)

    const labelColors = rows.reduce(
      (acc, idx) => {
        const key = (formDataObj[`lc:${idx}:key`] as string)?.trim()
        const color = (formDataObj[`lc:${idx}:color`] as string)?.trim()
        if (key && color) {
          acc[key] = color
        }
        return acc
      },
      {} as Record<string, string>,
    )

    onSubmit(labelColors)
  }

  const swatches = defaultColors?.length ? defaultColors.slice(0, 24) : SWATCH_COLORS

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="info">
          <Icon variant="PaletteIcon" size="24" />
          Edit label colors
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save label colors'
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
            {error?.error || 'Unable to update label colors'}
          </Banner>
        ) : null}

        <Text variant="subtext" theme="neutral">
          Assign colors to label keys. Colors are applied across all installs, components, and actions.
        </Text>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="label" weight="strong">
              Label colors
            </Text>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                const id = nextId.current++
                setRows((r) => [...r, id])
                setRowColors((c) => ({ ...c, [id]: swatches[rows.length % swatches.length] }))
              }}
            >
              <Icon variant="PlusIcon" size="16" />
              Add color
            </Button>
          </div>

          {rows.length === 0 && (
            <Text variant="subtext">No label colors configured</Text>
          )}

          {rows.map((idx) => (
            <fieldset
              key={idx}
              className="flex flex-col gap-2 border-t pt-2"
            >
              <div className="grid grid-cols-[1fr_auto_auto] gap-2 items-end">
                <label className="flex flex-col gap-1">
                  <Text variant="label">Label key</Text>
                  <Input
                    name={`lc:${idx}:key`}
                    type="text"
                    placeholder="e.g. env"
                    required
                    defaultValue={initialValues[`lc:${idx}:key`] || ''}
                  />
                </label>
                <label className="flex flex-col gap-1">
                  <Text variant="label">Color</Text>
                  <input
                    name={`lc:${idx}:color`}
                    type="color"
                    required
                    value={rowColors[idx] ?? '#3b82f6'}
                    onChange={(e) => setRowColors((c) => ({ ...c, [idx]: e.target.value }))}
                    className="h-9 w-12 rounded border border-cool-grey-300 dark:border-dark-grey-500 cursor-pointer"
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
                  <Icon variant="XIcon" size="16" />
                </Button>
              </div>
              <div className="flex flex-wrap gap-1">
                {swatches.map((color) => (
                  <button
                    key={color}
                    type="button"
                    className="w-5 h-5 rounded-sm border border-cool-grey-300 dark:border-dark-grey-500 cursor-pointer hover:scale-110 transition-transform"
                    style={{ backgroundColor: color }}
                    onClick={() => setRowColors((c) => ({ ...c, [idx]: color }))}
                    title={color}
                  />
                ))}
              </div>
            </fieldset>
          ))}
        </div>
      </form>
    </Modal>
  )
}
