import {
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useMediaQuery } from '../../hooks/use-media-query'
import {
  STATUS_THEME_COLOR,
  STATUS_THEME_ICON,
  type TSurfaceTheme,
} from '../../utils/status-theme'
import { Button } from './Button'
import { Card } from './Card'
import { Icon, type TIconVariant } from './Icon'
import { Text } from './Text'

export const BANNER_EXIT_MS = 200

const BANNER_FADE_MS = 150

const BANNER_SURFACE =
  'color-mix(in srgb, var(--banner-status) 12%, var(--card-bg))'

export type TBannerTheme = TSurfaceTheme

export interface IBanner extends HTMLAttributes<HTMLDivElement> {
  theme?: TBannerTheme
  heading?: ReactNode
  icon?: TIconVariant
  actions?: ReactNode
  dismissible?: boolean
  dismissLabel?: string
  onDismiss?: () => void
  className?: string
  children?: ReactNode
}

const stackGapStyle = (element: HTMLDivElement | null): CSSProperties => {
  const parent = element?.parentElement
  if (!element || !parent) return {}

  const gap = Number.parseFloat(getComputedStyle(parent).rowGap)
  if (!gap) return {}

  if (element.nextElementSibling) return { marginBottom: -gap }
  if (element.previousElementSibling) return { marginTop: -gap }
  return {}
}

export const Banner = ({
  theme = 'default',
  heading,
  icon,
  actions,
  dismissible = false,
  dismissLabel = 'Dismiss',
  onDismiss,
  className,
  children,
  style,
  ...props
}: IBanner) => {
  const reducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const rootRef = useRef<HTMLDivElement>(null)
  const removalTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const onDismissRef = useRef(onDismiss)
  const [dismissed, setDismissed] = useState(false)
  const [collapsedGap, setCollapsedGap] = useState<CSSProperties>({})
  const [removed, setRemoved] = useState(false)
  onDismissRef.current = onDismiss

  useEffect(
    () => () => {
      if (removalTimerRef.current) clearTimeout(removalTimerRef.current)
    },
    []
  )

  if (removed) return null

  const urgent = theme === 'error' || theme === 'warn'
  const hasHeading = heading !== undefined && heading !== null
  const hasBody = children !== undefined && children !== null

  const rootStyle = {
    '--banner-status': STATUS_THEME_COLOR[theme],
    transitionDuration: reducedMotion ? '1ms' : `${BANNER_EXIT_MS}ms`,
    ...(dismissed ? collapsedGap : {}),
    ...style,
  } as CSSProperties

  const surfaceStyle: CSSProperties = {
    backgroundColor: BANNER_SURFACE,
    transitionDuration: reducedMotion ? '1ms' : `${BANNER_FADE_MS}ms`,
  }

  const dismiss = () => {
    if (dismissed) return

    setCollapsedGap(stackGapStyle(rootRef.current))
    setDismissed(true)
    removalTimerRef.current = setTimeout(
      () => {
        setRemoved(true)
        onDismissRef.current?.()
      },
      reducedMotion ? 0 : BANNER_EXIT_MS
    )
  }

  return (
    <div
      ref={rootRef}
      role={urgent ? 'alert' : 'status'}
      aria-live={urgent ? 'assertive' : 'polite'}
      data-banner-theme={theme}
      className={cn(
        'grid transition-[grid-template-rows,margin] ease-in-out',
        dismissed ? 'grid-rows-[0fr]' : 'grid-rows-[1fr]',
        className
      )}
      style={rootStyle}
      {...props}
    >
      <div className="overflow-hidden">
        <Card
          padding="sm"
          style={surfaceStyle}
          className={cn(
            'flex items-start gap-2.5 transition-opacity ease-out',
            dismissed && 'opacity-0'
          )}
        >
          <span
            aria-hidden
            className="mt-1 flex shrink-0 text-[var(--banner-status)]"
          >
            <Icon variant={icon ?? STATUS_THEME_ICON[theme]} size={16} />
          </span>
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            {hasHeading ? (
              typeof heading === 'string' ? (
                <Text as="p" weight="medium">
                  {heading}
                </Text>
              ) : (
                heading
              )
            ) : null}
            {hasBody ? (
              typeof children === 'string' ? (
                <Text as="p" variant="caption" color="secondary">
                  {children}
                </Text>
              ) : (
                <div className="text-caption text-secondary">{children}</div>
              )
            ) : null}
            {actions ? (
              <div className="flex flex-wrap justify-end gap-2 pt-2">
                {actions}
              </div>
            ) : null}
          </div>
          {dismissible ? (
            <Button
              size="sm"
              variant="ghost"
              iconOnly
              aria-label={dismissLabel}
              onClick={dismiss}
            >
              <Icon variant="XIcon" size={16} />
            </Button>
          ) : null}
        </Card>
      </div>
    </div>
  )
}
