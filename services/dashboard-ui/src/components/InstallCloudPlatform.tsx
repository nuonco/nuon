'use client'

import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import { QuestionIcon } from '@phosphor-icons/react'

export const InstallPlatform: FC<{ platform: 'aws' | 'azure' | string }> = ({
  platform,
}) => {
  return (
    <span className="flex gap-2 items-center">
      {platform === 'azure' ? (
        <>
          <VscAzure className="text-md" /> {'Azure'}
        </>
      ) : platform === 'aws' ? (
        <>
          <FaAws className="text-xl mb-[-4px]" /> {'Amazon'}
        </>
      ) : (
        <>
          <QuestionIcon className="text-xl mb-[-4px]" /> {'Unknown'}
        </>
      )}
    </span>
  )
}
