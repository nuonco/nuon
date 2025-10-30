'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { deployComponents } from '@/actions/installs/deploy-components'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

// old stuff
import { CheckboxInput } from '@/components/old/Input'

export const DeployComponentButton = () => {
  return <Button></Button>
}

export const DeployComponent = () => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)

  const {
    data: deploysOk,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: deployComponents })

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_deploy',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: error?.error,
        },
      })
    }

    if (deploysOk) {
      trackEvent({
        event: 'components_deploy',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    }
  }, [deploysOk, error, headers])

  return (
    <Modal
      heading="Deploy all components"
      triggerButton={{
        children: 'Deploy all components',
        isMenuButton: true,
        variant: 'ghost',
      }}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to deploy components'}
          </Banner>
        ) : null}
        <Text variant="body">
          Are you sure you want to deploy components? This will deploy all
          components to this install.
        </Text>
        <CheckboxInput
          name="ack"
          defaultChecked={planOnly}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            setPlanOnly(Boolean(e?.currentTarget?.checked))
          }}
          labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
          labelText={'Plan Only?'}
        />
      </div>
    </Modal>
  )
}
