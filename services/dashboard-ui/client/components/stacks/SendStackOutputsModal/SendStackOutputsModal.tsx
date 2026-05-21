import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { CodeInput } from '@/components/common/form/CodeInput'
import type { TAPIError } from '@/types'

interface ISendStackOutputsModal extends IModal {
  phoneHomeId: string
  versionId: string
  onSend: (body: Record<string, unknown>) => void
  isPending: boolean
  error: TAPIError | undefined
}

export const SendStackOutputsModal = ({
  phoneHomeId,
  versionId,
  onSend,
  isPending,
  error,
  ...props
}: ISendStackOutputsModal) => {
  const [value, setValue] = useState('')

  const formatJson = () => {
    try {
      setValue(JSON.stringify(JSON.parse(value), null, 2))
    } catch {}
  }

  const isValidJson = (() => {
    try {
      JSON.parse(value)
      return true
    } catch {
      return false
    }
  })()

  const handleSubmit = () => {
    try {
      onSend(JSON.parse(value))
    } catch {
      // invalid JSON
    }
  }

  return (
    <Modal
      heading="Trigger phone home"
      size="lg"
      primaryActionTrigger={{
        children: isPending ? 'Sending...' : 'Trigger phone home',
        disabled: isPending || !isValidJson || !phoneHomeId,
        onClick: handleSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center gap-2">
          <Text variant="subtext">Version</Text>
          <Badge variant="code" size="md">{versionId}</Badge>
        </div>

        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between">
            <Text variant="body" weight="strong">Outputs (JSON)</Text>
            <Button
              variant="ghost"
              size="sm"
              onClick={formatJson}
              disabled={!isValidJson}
            >
              <Icon variant="BracketsCurlyIcon" size={14} />
              Format
            </Button>
          </div>
          <CodeInput
            language="json"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            error={!isValidJson && value.trim().length > 0}
            errorMessage="Invalid JSON"
            minHeight={200}
          />
        </div>

        {error && (
          <Text variant="subtext" theme="error">
            {error.error ?? 'Failed to send outputs'}
          </Text>
        )}
      </div>
    </Modal>
  )
}
