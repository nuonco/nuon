import { cn } from '@/utils/classnames'
import { Text, type IText } from './Text'

export interface ILabel extends Omit<IText, 'as'> {
  htmlFor?: string
}

export const Label = ({
  loading = false,
  loadingWidth,
  className,
  children,
  ...props
}: ILabel) => (
  <Text
    as="label"
    variant="label"
    weight="medium"
    loading={loading}
    loadingWidth={loadingWidth}
    className={cn('block text-secondary', className)}
    {...props}
  >
    {children}
  </Text>
)
