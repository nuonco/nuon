import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IEditLabelsModal extends Omit<IModal, 'onSubmit'> {
  labels: Record<string, string>
  isPending: boolean
  error: any
  onSave: (labels: Record<string, string>) => void
}

export const EditLabelsModal = ({
  labels: initialLabels,
  isPending,
  error,
  onSave,
  ...props
}: IEditLabelsModal) => {
  const [labels, setLabels] = useState<Record<string, string>>({ ...initialLabels })
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  const handleAdd = () => {
    const key = newKey.trim()
    if (!key) return
    setLabels((prev) => ({ ...prev, [key]: newValue.trim() }))
    setNewKey('')
    setNewValue('')
  }

  const handleRemove = (key: string) => {
    setLabels((prev) => {
      const next = { ...prev }
      delete next[key]
      return next
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAdd()
    }
  }

  const sortedKeys = Object.keys(labels).sort()

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
        onClick: () => onSave(labels),
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update labels'}
          </Banner>
        ) : null}

        {sortedKeys.length > 0 ? (
          <div className="flex flex-col gap-2">
            {sortedKeys.map((key) => (
              <div key={key} className="flex items-center gap-2">
                <Text
                  variant="body"
                  className="flex-1 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded text-sm truncate font-mono"
                >
                  {key}: {labels[key]}
                </Text>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleRemove(key)}
                  aria-label={`Remove label ${key}`}
                >
                  <Icon variant="X" size={14} />
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <Text variant="body" theme="neutral">
            No labels set.
          </Text>
        )}

        <hr />

        <div className="flex items-end gap-2">
          <div className="flex-1">
            <Text variant="label" theme="neutral" className="mb-1 block">
              Key
            </Text>
            <input
              type="text"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="e.g. env"
              className="w-full px-3 py-1.5 border rounded-md text-sm bg-transparent border-gray-300 dark:border-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div className="flex-1">
            <Text variant="label" theme="neutral" className="mb-1 block">
              Value
            </Text>
            <input
              type="text"
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="e.g. production"
              className="w-full px-3 py-1.5 border rounded-md text-sm bg-transparent border-gray-300 dark:border-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <Button variant="secondary" size="sm" onClick={handleAdd} disabled={!newKey.trim()}>
            <Icon variant="PlusIcon" size={14} />
            Add
          </Button>
        </div>
      </div>
    </Modal>
  )
}
