import type { AnchorHTMLAttributes, HTMLAttributes } from 'react'
import type { TCompositeStatus } from '@/types/ctl-api.types'
import { cn } from '@/utils/classnames'
import { Card } from '../atoms/Card'
import { Link } from '../atoms/Link'
import { Status } from '../atoms/Status'

export interface IGraphNodeCard
  extends Omit<HTMLAttributes<HTMLDivElement>, 'color'> {
  href?: string
  selected?: boolean
  status?: string | TCompositeStatus
}

const LINK_RESET =
  'block h-full w-full rounded-xl text-inherit no-underline decoration-transparent outline-none ' +
  'hover:text-inherit hover:decoration-transparent active:text-inherit ' +
  'focus-visible:rounded-xl focus-ring'

const SELECTED_CLASSES = 'outline outline-2 outline-offset-2 outline-accent'

export const GraphNodeCard = ({
  href,
  selected = false,
  status,
  className,
  children,
  ...props
}: IGraphNodeCard) => {
  const body = (
    <Card
      padding="sm"
      interactive={Boolean(href)}
      className={cn(
        'flex h-full min-h-0 w-full min-w-0 items-start gap-2',
        !href && selected && SELECTED_CLASSES,
        !href && className
      )}
      {...(!href ? props : {})}
    >
      {status ? <Status status={status} variant="icon" /> : null}
      <div className="flex min-w-0 flex-1 flex-col gap-1">{children}</div>
    </Card>
  )

  if (!href) return body

  return (
    <Link
      href={href}
      aria-current={selected || undefined}
      className={cn(LINK_RESET, selected && SELECTED_CLASSES, className)}
      {...(props as unknown as AnchorHTMLAttributes<HTMLAnchorElement>)}
    >
      {body}
    </Link>
  )
}
