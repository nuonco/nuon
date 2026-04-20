import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { getInstallStack, postPhoneHome } from '@/lib'
import { PhoneHomeModal, type PhoneHomeResult } from './PhoneHomeModal'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IPhoneHomeModalContainer extends IModal {
  installId: string
  orgId: string
}

export const PhoneHomeModalContainer = ({
  installId,
  orgId,
  ...props
}: IPhoneHomeModalContainer) => {
  const [result, setResult] = useState<PhoneHomeResult | undefined>(undefined)

  const { data: stack, isLoading: isLoadingStack } = useQuery({
    queryKey: ['install-stack', installId],
    queryFn: () => getInstallStack({ installId, orgId }),
    enabled: !!installId && !!orgId,
  })

  const phoneHomeId = stack?.versions?.[0]?.phone_home_id

  const { mutate: sendPhoneHome, isPending: isSubmitting } = useMutation({
    mutationFn: (body: Record<string, unknown>) =>
      postPhoneHome({ installId, orgId, phoneHomeId: phoneHomeId!, body }),
    onMutate: () => {
      setResult(undefined)
    },
    onSuccess: (data) => {
      setResult({ status: 'success', data })
    },
    onError: (err: TAPIError) => {
      setResult({ status: 'error', error: err })
    },
  })

  return (
    <PhoneHomeModal
      installId={installId}
      stack={stack}
      isLoadingStack={isLoadingStack}
      onSendPhoneHome={(body) => sendPhoneHome(body)}
      isSubmitting={isSubmitting}
      result={result}
      {...props}
    />
  )
}
