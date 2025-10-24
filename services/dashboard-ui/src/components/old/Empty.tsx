import classNames from 'classnames'
import React, { type FC } from 'react'
import Image from 'next/image'
import { Text } from '@/components/old/Typography'

type TEmptyVariant =
  | '404'
  | 'actions'
  | 'diagram'
  | 'history'
  | 'search'
  | 'table'

interface IEmptyGraphic {
  isDarkModeOnly?: boolean
  isSmall?: boolean
  variant?: TEmptyVariant
}

export const EmptyGraphic: FC<IEmptyGraphic> = ({
  isDarkModeOnly = false,
  isSmall = false,
  variant = '404',
}) => {
  return (
    <>
      <Image
        className={classNames('w-auto relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src={`/empty-state/${variant}-light${isSmall ? '-small' : ''}.svg`}
        alt=""
        height={90}
        width={150}
      />
      <Image
        className={classNames('w-auto relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src={`/empty-state/${variant}-dark${isSmall ? '-small' : ''}.svg`}
        alt=""
        height={90}
        width={150}
      />
    </>
  )
}

interface IEmpty extends React.HTMLAttributes<HTMLDivElement> {
  emptyTitle?: string
  emptyMessage?: string
  isDarkModeOnly?: boolean
  isSmall?: boolean
  variant?: TEmptyVariant
}

export const Empty: FC<IEmpty> = ({
  emptyMessage = 'Nothing found',
  emptyTitle = 'Nothing to show',
  isDarkModeOnly = false,
  isSmall = false,
  variant,
}) => {
  return (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyGraphic
        variant={variant}
        isDarkModeOnly={isDarkModeOnly}
        isSmall={isSmall}
      />
      <Text className="mt-6" variant="med-14">
        {emptyTitle}
      </Text>
      <Text variant="reg-12" className="text-center">
        {emptyMessage}
      </Text>
    </div>
  )
}
