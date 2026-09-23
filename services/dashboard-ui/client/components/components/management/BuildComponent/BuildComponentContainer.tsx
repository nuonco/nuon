import { useLocation, useNavigate } from 'react-router'
import { useContext, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/common/Badge'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { AppContext } from '@/providers/app-provider'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { buildComponent } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type { TComponent } from '@/types'
import {
  BuildComponentButton as BuildComponentButtonComponent,
  BuildComponentModal,
} from './BuildComponent'

export const BuildComponentButtonContainer = ({
  component,
  onClick: _onClick,
  redirectOnSuccess,
  ...props
}: IButtonAsButton & {
  component: TComponent
  redirectOnSuccess?: boolean
}) => {
  const { addModal } = useSurfaces()
  const modal = (
    <BuildComponentModalContainer
      component={component}
      redirectOnSuccess={redirectOnSuccess}
    />
  )
  return (
    <BuildComponentButtonComponent onClick={() => addModal(modal)} {...props} />
  )
}

export const BuildComponentModalContainer = ({
  component,
  redirectOnSuccess = true,
  ...props
}: IModal & {
  component: TComponent
  redirectOnSuccess?: boolean
}) => {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const appId = useContext(AppContext)?.app?.id ?? component.app_id
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const {
    data: build,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: () =>
      buildComponent({ componentId: component.id, orgId: org.id }),
    onSuccess: (build) => {
      addToast(
        <Toast heading="Build started" theme="info">
          <Text>
            Building{' '}
            <Badge variant="code" size="md">
              {component.name}
            </Badge>
            . This may take a few minutes.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
      if (redirectOnSuccess && build?.id) {
        navigate(`${pathname}/builds/${build.id}`)
      }
    },
    onError: () => {
      addToast(
        <Toast heading="Build failed" theme="error">
          <Text>
            Unable to build{' '}
            <Badge variant="code" size="md">
              {component.name}
            </Badge>
            .
          </Text>
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
        props: { orgId: org.id, appId, componentId: component.id },
      })
    }
    if (build) {
      trackEvent({
        event: 'component_build',
        status: 'ok',
        user,
        props: { orgId: org.id, appId, componentId: component.id },
      })
    }
  }, [build, error, org.id, appId, component.id, user])

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
