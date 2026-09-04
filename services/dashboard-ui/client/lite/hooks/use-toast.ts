import { useContext } from 'react'
import { ToastContext } from '../providers/toast-provider'

export const useToast = () => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used inside ToastProvider')
  }
  return context
}
