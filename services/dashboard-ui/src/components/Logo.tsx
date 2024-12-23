import classNames from 'classnames'
import React, { type FC } from 'react'
import Image from 'next/image'

export const Logo: FC<{ isDarkModeOnly?: boolean }> = ({
  isDarkModeOnly = false,
}) => {
  return (
    <a href="/">
      <span className="sr-only">Nuon</span>
      <Image
        className={classNames('w-auto h-8 relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/light.svg"
        alt="nuon logo"
        height={32}
        width={110}
      />
      <Image
        className={classNames('w-auto h-8 relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/dark.svg"
        alt="nuon logo"
        height={32}
        width={110}
      />
    </a>
  )
}
