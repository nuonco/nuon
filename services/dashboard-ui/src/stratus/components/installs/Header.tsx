'use client'

import React, { type FC } from 'react'
import { Header } from '@/stratus/components/dashboard'
import { InstallHeaderDetails } from './HeaderDetails'
import { InstallHeaderGroup } from './HeaderGroup'

export const InstallHeader: FC = () => {
  return (
    <Header className="border-b">
      <InstallHeaderGroup />
      <InstallHeaderDetails />
    </Header>
  )
}
