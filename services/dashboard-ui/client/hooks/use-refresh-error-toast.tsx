import { useEffect } from 'react'
import { useToast } from '@/hooks/use-toast'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { TAPIError } from '@/types'

export function useRefreshErrorToast(error: unknown, data: unknown) {
  const { addToast } = useToast()

  useEffect(() => {
    if (error && data) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])
}
