'use client'

import { useEffect } from 'react'
import { Toast, TToastTheme } from '@/components/surfaces/Toast'
import { useToast } from './use-toast'

export interface IUseServerActionToast {
  data: any
  error: any
  successHeading?: React.ReactNode
  successContent?: React.ReactNode
  errorHeading?: React.ReactNode
  errorContent?: React.ReactNode
  successTheme?: TToastTheme
  errorTheme?: TToastTheme
  onSuccess?: () => void
  onError?: () => void
}

export function useServerActionToast({
  data,
  error,
  successHeading = 'Success!',
  successContent = <span>Operation completed successfully.</span>,
  errorHeading = 'Error',
  errorContent = <span>Something went wrong.</span>,
  successTheme = 'info',
  errorTheme = 'error',
  onSuccess,
  onError,
}: IUseServerActionToast) {
  const { addToast } = useToast()

  useEffect(() => {
    if (data && !error) {
      addToast(
        <Toast theme={successTheme} heading={successHeading}>
          {successContent}
        </Toast>
      )
      if (onSuccess) onSuccess()
    }
    if (!data && error) {
      addToast(
        <Toast theme={errorTheme} heading={errorHeading}>
          {errorContent}
        </Toast>
      )
      if (onError) onError()
    }
  }, [data, error])
}
