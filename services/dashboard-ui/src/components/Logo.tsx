import React, { type FC } from 'react'
import Image from 'next/image'

export const Logo: FC = () => {
  return (
    <a href="/">
      <span className="sr-only">Nuon</span>
      <Image
        className="w-auto h-8 relative block dark:hidden"
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/light.svg"
        alt="nuon logo"
        height={32}
        width={110}
      />
      <Image
        className="w-auto h-8 relative hidden dark:block"
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/dark.svg"
        alt="nuon logo"
        height={32}
        width={110}
      />
    </a>
  )
}
