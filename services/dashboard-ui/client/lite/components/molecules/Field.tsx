import {
  cloneElement,
  isValidElement,
  useId,
  type HTMLAttributes,
  type ReactElement,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Label } from '../atoms/Label'
import { Text } from '../atoms/Text'

export interface IField
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  label?: ReactNode
  description?: ReactNode
  error?: ReactNode
  optional?: boolean
  loading?: boolean
  loadingWidth?: number
  children: ReactElement
}

type TControlProps = {
  id?: string
  'aria-invalid'?: boolean
  'aria-describedby'?: string
}

export const Field = ({
  label,
  description,
  error,
  optional = false,
  loading = false,
  loadingWidth,
  className,
  children,
  ...props
}: IField) => {
  const generatedId = useId()
  const child = isValidElement<TControlProps>(children) ? children : null
  const controlId = child?.props.id ?? `${generatedId}-control`
  const descriptionId = description ? `${generatedId}-description` : undefined
  const errorId = error ? `${generatedId}-error` : undefined
  const describedBy =
    [child?.props['aria-describedby'], descriptionId, errorId]
      .filter(Boolean)
      .join(' ') || undefined

  return (
    <div className={cn('flex min-w-0 flex-col gap-1.5', className)} {...props}>
      {label ? (
        <span className="flex items-baseline justify-between gap-3">
          <Label
            htmlFor={controlId}
            loading={loading}
            loadingWidth={loadingWidth}
          >
            {label}
          </Label>
          {optional ? (
            <Text variant="caption" color="tertiary">
              Optional
            </Text>
          ) : null}
        </span>
      ) : null}
      {description ? (
        <Text id={descriptionId} variant="caption" color="tertiary">
          {description}
        </Text>
      ) : null}
      {child
        ? cloneElement(child, {
            id: controlId,
            'aria-invalid': child.props['aria-invalid'] || !!error || undefined,
            'aria-describedby': describedBy,
          })
        : children}
      {error ? (
        <Text id={errorId} variant="caption" className="text-field-invalid">
          {error}
        </Text>
      ) : null}
    </div>
  )
}
