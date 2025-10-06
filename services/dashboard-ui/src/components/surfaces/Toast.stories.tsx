import { ToastProvider } from '@/providers/toast-provider'
import { useToast } from '@/hooks/use-toast'
import { Button } from '@/components/common/Button'
import { Toast } from './Toast'

const ToastTrigger = ({ theme, children }) => {
  const { addToast } = useToast()
  return (
    <Button
      onClick={() => {
        addToast(
          <Toast theme={theme} heading={`${theme} toast`}>
            This is a {theme} toast.
          </Toast>
        )
      }}
    >
      {children}
    </Button>
  )
}

export const Default = () => (
  <ToastProvider>
    <ToastTrigger theme="default">Default toast</ToastTrigger>
  </ToastProvider>
)

export const Themes = () => (
  <ToastProvider>
    <div className="flex gap-4">
      <ToastTrigger theme="default">Default</ToastTrigger>
      <ToastTrigger theme="info">Info</ToastTrigger>
      <ToastTrigger theme="success">Success</ToastTrigger>
      <ToastTrigger theme="warn">Warn</ToastTrigger>
      <ToastTrigger theme="error">Error</ToastTrigger>
    </div>
  </ToastProvider>
)
