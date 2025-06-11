'use client'

import React, { type FC, useEffect, useState } from 'react'
import {
  ArrowURightDown,
  ArrowURightUp,
  CodeBlock,
  ListChecks,
  LockLaminated,
  PencilSimpleLine,
  StackMinus,
  Trash,
} from '@phosphor-icons/react/dist/ssr'
import { Button, Dropdown, Menu, Text } from '@/stratus/components/common'

export const InstallManageDropdown: FC = () => {
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
          Edit inputs <PencilSimpleLine />
        </Button>
        <Button>
          Auto approve changes <ListChecks />
        </Button>
        <Button>
          View state <CodeBlock />
        </Button>
        <hr />
        <Text variant="label" theme="muted">
          Controlls
        </Text>
        <Button>
          Reprovision install <ArrowURightUp />
        </Button>
        <Button>
          Deprovision install <ArrowURightDown />
        </Button>
        <Button>
          Deprovision stack <StackMinus />
        </Button>
        <hr />
        <Text variant="label" theme="muted">
          Danger
        </Text>
        <Button>
          Break glass permissions <LockLaminated />
        </Button>
        <Button className="!text-red-800 dark:!text-red-500">
          Forget install
          <Trash />
        </Button>
      </Menu>
    </Dropdown>
  )
}
