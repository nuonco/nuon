import { useLocation, useNavigate } from 'react-router'
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
import { buildComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { BuildComponentButton as BuildComponentButtonComponent, BuildComponentModal } from './BuildComponent'

export const BuildComponentButtonContainer = ({
  component,
  onClick: _onClick,
  ...props
}: IButtonAsButton & {
  component: TComponent
}) => {
  const { addModal } = useSurfaces()
  const modal = <BuildComponentModalContainer component={component} />
  return (
    <BuildComponentButtonComponent
      onClick={() => addModal(modal)}
      {...props}
    />
  )
}

const BuildComponentModalContainer = ({
  component,
  ...props
}: IModal & {
  component: TComponent
}) => {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { data: build, error, mutate, isPending: isLoading } = useMutation({
    mutationFn: () => buildComponent({ componentId: component.id, orgId: org.id }),
    onSuccess: (build) => {
      addToast(
        <Toast heading={`${component.name} build started`} theme="success">
          <Text>Build for {component.name} was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      if (build?.id) {
        navigate(`${pathname}/builds/${build.id}`)
      }
    },
    onError: () => {
      addToast(
        <Toast heading="Component build failed" theme="error">
          <Text>Unable to build {component.name} component.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_build',
        status: 'error',
        user,
        props: { orgId: org.id, appId: app.id, componentId: component.id },
      })
    }
    if (build) {
      trackEvent({
        event: 'component_build',
        status: 'ok',
        user,
        props: { orgId: org.id, appId: app.id, componentId: component.id },
      })
    }
  }, [build, error, org.id, app.id, component.id, user])

  return (
    <BuildComponentModal
      component={component}
      isLoading={isLoading}
      error={error}
      onBuild={() => mutate()}
      {...props}
    />
  )
}
