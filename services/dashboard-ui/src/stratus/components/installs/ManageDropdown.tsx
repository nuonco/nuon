'use client'

import React, { useEffect, useState } from 'react'
import { Button, Dropdown, Icon, Menu, Text } from '@/stratus/components/common'

export const InstallManageDropdown = () => {
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => setIsMobile(window.innerWidth < 768)
    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  return (
    <Dropdown
      id="install-manage"
      buttonText="Manage"
      alignment={isMobile ? 'left' : 'right'}
      variant="primary"
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="muted">
          Settings
        </Text>
        <Button>
          Edit inputs <Icon variant="PencilSimpleLine" />
        </Button>
        <Button>
          Auto approve changes <Icon variant="ListChecks" />
        </Button>
        <Button>
          View state <Icon variant="CodeBlock" />
        </Button>
        <hr />
        <Text variant="label" theme="muted">
          Controlls
        </Text>
        <Button>
          Reprovision install <Icon variant="ArrowURightUp" />
        </Button>
        <Button>
          Deprovision install <Icon variant="ArrowURightDown" />
        </Button>
        <Button>
          Deprovision stack <Icon variant="StackMinus" />
        </Button>
        <hr />
        <Text variant="label" theme="muted">
          Danger
        </Text>
        <Button>
          Break glass permissions <Icon variant="LockLaminated" />
        </Button>
        <Button className="!text-red-800 dark:!text-red-500">
          Forget install
          <Icon variant="Trash" />
        </Button>
      </Menu>
    </Dropdown>
  )
}
