import classNames from 'classnames'
import React, { type FC } from 'react'
import Image from 'next/image'

export const EmptyStateGraphic: FC<{ isDarkModeOnly?: boolean }> = ({
  isDarkModeOnly = false,
}) => {
  return (
    <>
      <Image
        className={classNames('w-auto relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src="/empty-diagram-light.svg"
        alt=""
        height={90}
        width={150}
      />
      <Image
        className={classNames('w-auto relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src="/empty-diagram-dark.svg"
        alt=""
        height={90}
        width={150}
      />
    </>
  )
}
