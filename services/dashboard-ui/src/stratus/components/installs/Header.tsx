'use client'

import React from 'react'
import { Header } from '@/stratus/components/dashboard'
import { InstallHeaderDetails } from './HeaderDetails'
import { InstallHeadingGroup } from './HeadingGroup'

export const InstallHeader = () => {
  return (
    <Header className="border-b">
      <InstallHeadingGroup />
      <InstallHeaderDetails />
    </Header>
  )
}
