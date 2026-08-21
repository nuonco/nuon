export default {
  title: 'Consent/ConsentToast',
}

import { ToastProvider } from '@/providers/toast-provider'
import { ConsentToast } from './ConsentToast'

export const Default = () => (
  <ToastProvider>
    <div className="relative h-72 w-full">
      <ConsentToast onAccept={() => {}} onDecline={() => {}} />
    </div>
  </ToastProvider>
)
