'use client'

import React, { useState } from 'react'

interface TabProps {
  children: React.ReactNode
  title: string
}

interface TabsProps {
  children: React.ReactNode
  defaultTab?: number
}

export const Tabs = ({ children, defaultTab = 0 }: TabsProps): JSX.Element => {
  const [activeTab, setActiveTab] = useState<number>(defaultTab)

  // Filter out only Tab components from children
  // const tabs = React.Children.toArray(children).filter(isTabElement)
  const tabs = React.Children.toArray(children)

  return (
    <div className="w-full max-w-3xl mx-auto">
      {/* Tab navigation */}
      <div className="flex border-b border-gray-200">
        {tabs.map((tab: any, index) => (
          <button
            key={index}
            className={`px-4 py-2 text-sm font-medium transition-colors duration-200 ${
              activeTab === index
                ? 'text-blue-600 border-b-2 border-blue-600'
                : 'text-gray-500 hover:text-gray-700'
            }`}
            onClick={() => setActiveTab(index)}
          >
            {tab.props.title}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="p-4">{tabs[activeTab]}</div>
    </div>
  )
}

export const Tab = ({ children }: TabProps): JSX.Element => {
  return <div>{children}</div>
}
