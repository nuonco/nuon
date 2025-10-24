'use client'

import { useSearchParams, useRouter, usePathname } from 'next/navigation'
import React, { useState } from 'react'
import { slugifyString, sentanceCase } from '@/utils'

interface TabProps {
  children: React.ReactNode
  title: string
}

interface TabsProps {
  children: React.ReactNode
  defaultTab?: number
}

export const Tabs = ({ children, defaultTab = 0 }: TabsProps): JSX.Element => {
  const path = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const queryActiveTab = searchParams.get('terraform')
  const tabs = React.Children.toArray(children)
  const queryTab = tabs?.findIndex(
    (t) =>
      (t as React.ReactElement)?.props?.title ===
      sentanceCase(queryActiveTab || '')
  )
  const [activeTab, setActiveTab] = useState<number>(
    queryTab !== -1 ? queryTab : defaultTab
  )

  return (
    <div className="w-full flex flex-col gap-4">
      {/* Tab navigation */}
      <div className="flex border-b">
        {tabs.map((tab: any, index) => (
          <button
            key={index}
            className={`px-4 py-2 text-base border-b font-medium transition-colors duration-200 !shadow-none ${
              activeTab === index
                ? 'text-primary-600 dark:text-primary-400 border-current'
                : 'text-cool-grey-600 dark:text-cool-grey-400 border-transparent'
            }`}
            onClick={() => {
              router.push(
                `${path}?${new URLSearchParams({ terraform: slugifyString(tab?.props?.title) }).toString()}`
              )
              setActiveTab(index)
            }}
          >
            {tab.props.title}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div>{tabs[activeTab]}</div>
    </div>
  )
}

export const Tab = ({ children }: TabProps): JSX.Element => {
  return <div>{children}</div>
}
