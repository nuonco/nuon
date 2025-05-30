'use client'

import classNames from 'classnames'
import { usePathname } from 'next/navigation'
import React, { type FC, useRef, useState } from 'react'
import { SplitHorizontal } from '@phosphor-icons/react'
import { Link, Text, Tooltip } from '@/stratus/components/common'
import { usePage } from '@/stratus/context'
import type { TNavLink } from '@/types'
import './PageNav.css'

interface IPageNav {
  basePath: string
  links: Array<TNavLink>
}

export const PageNav: FC<IPageNav> = ({ basePath, links }) => {
  const { isPageNavOpen, closePageNav, openPageNav, togglePageNav } = usePage()
  const [dragging, setDragging] = useState(false)
  const handleRef = useRef<HTMLDivElement>(null)
  const startXRef = useRef<number | null>(null)

  const handleDragStart = (e: React.MouseEvent | React.TouchEvent) => {
    setDragging(true)
    const startX = 'touches' in e ? e.touches[0].clientX : e.clientX
    startXRef.current = startX
  }

  const handleDragMove = (e: React.MouseEvent | React.TouchEvent) => {
    if (!dragging || startXRef.current === null) return

    const currentX = 'touches' in e ? e.touches[0].clientX : e.clientX
    const deltaX = currentX - startXRef.current

    if (deltaX < -1 && isPageNavOpen) {
      closePageNav()
      setDragging(false)
    } else if (deltaX > 1 && !isPageNavOpen) {
      openPageNav()
      setDragging(false)
    }
  }

  const handleDragEnd = () => {
    setDragging(false)
    startXRef.current = null
  }

  return (
    <aside
      className={classNames('page-nav', {
        'is-open': isPageNavOpen,
      })}
    >
      <nav>
        {links.map((link) => (
          <PageNavLink key={link.path} basePath={basePath} {...link} />
        ))}
      </nav>
      <div
        ref={handleRef}
        className="page-nav-handle"
        onMouseDown={handleDragStart}
        onMouseMove={handleDragMove}
        onMouseUp={handleDragEnd}
        onTouchStart={handleDragStart}
        onTouchMove={handleDragMove}
        onTouchEnd={handleDragEnd}
      >
        <button
          className="page-nav-handle-button"
          onClick={() => {
            togglePageNav()
          }}
        >
          <SplitHorizontal />
        </button>
      </div>
    </aside>
  )
}

const PageNavLink: FC<TNavLink & { basePath: string }> = ({
  basePath,
  icon,
  path,
  text,
}) => {
  const { isPageNavOpen } = usePage()
  const pathName = usePathname()
  const normalizePath = (path: string) =>
    path.endsWith('/') ? path.slice(0, -1) : path
  const normalizedPathName = normalizePath(pathName)
  const fullPath = normalizePath(`${basePath}${path}`)
  const isActive =
    fullPath === normalizedPathName ||
    (path !== `/` && normalizedPathName.startsWith(`${fullPath}/`))

  const link = (
    <Link
      aria-current={isActive ? 'page' : undefined}
      href={`${basePath}/${path}`}
      isActive={isActive}
      variant="nav"
    >
      <span>{icon}</span>
      <span className="link-text">{text}</span>
    </Link>
  )

  return isPageNavOpen ? (
    link
  ) : (
    <Tooltip
      position="right"
      tipContent={
        <Text variant="subtext" weight="stronger">
          {text
            .trim()
            .split(' ')
            .at(-1)
            ?.replace(/^./, (char) => char.toUpperCase())}
        </Text>
      }
    >
      {link}
    </Tooltip>
  )
}
