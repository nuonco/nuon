import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { buildComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { BuildAllComponentsButton as BuildAllComponentsButtonComponent, BuildAllComponentsModal } from './BuildAllComponents'

export const BuildAllComponentsButtonContainer = ({ onClick: _onClick, ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <BuildAllComponentsModalContainer />
  return (
    <BuildAllComponentsButtonComponent
      onClick={() => addModal(modal)}
      {...props}
    />
  )
}

export const BuildAllComponentsModalContainer = ({ ...props }: IModal) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { data: buildOk, error, mutate, isPending: isLoading } = useMutation({
    mutationFn: () => buildComponents({ appId: app.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading={`${app.name} component builds started`} theme="success">
          <Text>Build all components workflow was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Component builds failed" theme="error">
          <Text>Unable to build all components for {app.name}.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_build',
        status: 'error',
        user,
        props: { appId: app.id, orgId: org.id, err: error?.error },
      })
    }
    if (buildOk) {
      trackEvent({
        event: 'components_build',
        status: 'ok',
        user,
        props: { appId: app.id, orgId: org.id },
      })
    }
  }, [buildOk, error, app.id, org.id, user])

  return (
    <BuildAllComponentsModal
      appName={app.name}
      isLoading={isLoading}
      error={error}
      onBuild={() => mutate()}
      {...props}
    />
  )
}
