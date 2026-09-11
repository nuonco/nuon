import { useCallback } from 'react'
import { useToast } from '@/hooks/use-toast'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'

export function useRefreshErrorToast() {
  const { addToast } = useToast()

  return useCallback(
    (message: string) => {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{message}</Text>
        </Toast>
      )
    },
    [addToast]
  )
}
