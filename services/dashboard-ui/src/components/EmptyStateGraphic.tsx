import classNames from 'classnames'
import React, { type FC } from 'react'
import Image from 'next/image'

interface IEmptyStateGraphic {
  isDarkModeOnly?: boolean
  variant?: "404" | "actions" | "diagram" | "history" | "search" | "table"
}

export const EmptyStateGraphic: FC<IEmptyStateGraphic> = ({
  isDarkModeOnly = false,
  variant = "404",
}) => {
  return (
    <>
      <Image
        className={classNames('w-auto relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src={`/empty-state/${variant}-light.svg`}
        alt=""
        height={90}
        width={150}
      />
      <Image
        className={classNames('w-auto relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src={`/empty-state/${variant}-dark.svg`}
        alt=""
        height={90}
        width={150}
      />
    </>
  )
}
