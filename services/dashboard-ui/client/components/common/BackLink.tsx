import { useNavigate } from 'react-router'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text, type IText } from './Text'

interface IBackLink extends IText {}

export const BackLink = ({
  className,
  children = (
    <>
      <Icon variant="CaretLeftIcon" weight="bold" /> Back
    </>
  ),
  variant = 'base',
  weight = 'strong',
  ...props
}: IBackLink) => {
  const navigate = useNavigate()

  return (
    <Text
      className={cn(
        '!flex items-center gap-1.5 cursor-pointer w-fit',
        'text-link',
        'hover:text-link-hover',
        'focus:text-link-hover',
        'active:text-link-active',
        'focus-visible:rounded',
        className
      )}
      onClick={() => {
        navigate(-1)
      }}
      variant={variant}
      weight={weight}
      {...props}
    >
      {children}
    </Text>
  )
}
