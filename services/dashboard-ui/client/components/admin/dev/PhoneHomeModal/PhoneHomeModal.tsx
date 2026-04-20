import { useState } from 'react'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Textarea } from '@/components/common/form/Textarea'
import type { TInstallStack, TAPIError } from '@/types'

type PhoneHomeResult =
  | { status: 'success'; data: string }
  | { status: 'error'; error: TAPIError }

interface IPhoneHomeModal extends IModal {
  installId: string
  stack: TInstallStack | undefined
  isLoadingStack: boolean
  onSendPhoneHome: (body: Record<string, unknown>) => void
  isSubmitting: boolean
  result: PhoneHomeResult | undefined
}

export const PhoneHomeModal = ({
  installId,
  stack,
  isLoadingStack,
  onSendPhoneHome,
  isSubmitting,
  result,
  ...props
}: IPhoneHomeModal) => {
  const latestVersion = stack?.versions?.[0]
  const latestRun = latestVersion?.runs?.[0]
  const phoneHomeId = latestVersion?.phone_home_id
  const latestBody = latestRun?.data_contents ?? latestRun?.data ?? {}

  const [editorValue, setEditorValue] = useState<string | undefined>(undefined)

  const displayValue =
    editorValue !== undefined
      ? editorValue
      : JSON.stringify(latestBody, null, 2)

  const handleSubmit = () => {
    try {
      const parsed = JSON.parse(displayValue)
      onSendPhoneHome(parsed)
    } catch {
      // invalid JSON, don't submit
    }
  }

  const isValidJson = (() => {
    try {
      JSON.parse(displayValue)
      return true
    } catch {
      return false
    }
  })()

  return (
    <Modal
      heading="Phone home"
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Sending...' : 'Send phone home',
        disabled: isSubmitting || isLoadingStack || !phoneHomeId || !isValidJson,
        onClick: handleSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center gap-2">
          <Text variant="subtext">Install</Text>
          <Badge variant="code">{installId}</Badge>
        </div>

        {phoneHomeId && (
          <div className="flex flex-wrap items-center gap-2">
            <Text variant="subtext">Phone home ID</Text>
            <Badge variant="code">{phoneHomeId}</Badge>
          </div>
        )}

        {isLoadingStack && (
          <Text variant="subtext" className="text-gray-500">
            Loading latest phone home data...
          </Text>
        )}

        {!isLoadingStack && !latestVersion && (
          <Text variant="subtext" className="text-gray-500">
            No stack version found for this install.
          </Text>
        )}

        {!isLoadingStack && latestVersion && (
          <>
            {latestRun?.created_at && (
              <Text variant="subtext" className="text-gray-500">
                Last phone home: {new Date(latestRun.created_at).toLocaleString()}
              </Text>
            )}

            <div className="flex flex-col gap-1">
              <Textarea
                labelProps={{ labelText: 'Request body' }}
                value={displayValue}
                onChange={(e) => setEditorValue(e.target.value)}
                placeholder='{"request_type": "Update"}'
                wrap="off"
                spellCheck={false}
                className="font-mono !whitespace-pre"
                style={{ minHeight: 300, maxHeight: 500, resize: 'vertical' }}
              />
              {!isValidJson && editorValue !== undefined && (
                <Text variant="subtext" className="text-red-500">
                  Invalid JSON
                </Text>
              )}
            </div>
          </>
        )}

        {result && (
          <div
            className={`rounded-md border p-4 ${
              result.status === 'success'
                ? 'border-green-300 bg-green-50 dark:border-green-700 dark:bg-green-950'
                : 'border-red-300 bg-red-50 dark:border-red-700 dark:bg-red-950'
            }`}
          >
            <div className="flex flex-col gap-2">
              <div className="flex items-center gap-2">
                <span
                  className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-mono font-semibold ${
                    result.status === 'success'
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                      : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                  }`}
                >
                  {result.status === 'success'
                    ? '201 Created'
                    : `${result.error.status ?? 'Error'}`}
                </span>
                <Text
                  variant="base"
                  weight="strong"
                  className={
                    result.status === 'success'
                      ? 'text-green-700 dark:text-green-300'
                      : 'text-red-700 dark:text-red-300'
                  }
                >
                  {result.status === 'success' ? 'Success' : 'Failed'}
                </Text>
              </div>

              {result.status === 'success' ? (
                <Text
                  variant="subtext"
                  className="font-mono text-green-700 dark:text-green-300"
                >
                  {JSON.stringify(result.data)}
                </Text>
              ) : (
                <div className="flex flex-col gap-1">
                  <Text
                    variant="subtext"
                    className="font-mono text-red-700 dark:text-red-300"
                  >
                    {result.error.error}
                  </Text>
                  {result.error.description && (
                    <Text variant="subtext" className="text-red-600 dark:text-red-400">
                      {result.error.description}
                    </Text>
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}

export type { PhoneHomeResult }
