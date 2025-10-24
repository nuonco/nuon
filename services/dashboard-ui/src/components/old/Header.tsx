import React, { type FC } from 'react'
import { Link } from '@/components/old/Link'
import { Logo } from '@/components/old/Logo'
import { SignOutButton } from '@/components/old/Profile'

// TODO: maybe LayoutHeader?
export const Header: FC = () => {
  return (
    <header className="flex flex-wrap items-center justify-between gap-6 pb-6 border-b">
      <div className="flex items-center gap-6">
        <Logo />
      </div>

      <div className="flex gap-4 items-center">
        <Link className="text-sm" href="https://docs.nuon.co" target="_blank">
          Documentation
        </Link>

        <SignOutButton />
      </div>
    </header>
  )
}
