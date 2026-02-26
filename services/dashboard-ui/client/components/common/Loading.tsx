import { cn } from '@/utils/classnames'

export const Loading = ({
  strokeWidth = 'default',
  variant = 'default',
}: {
  strokeWidth?: 'default' | 'thick'
  variant?: 'default' | 'large'
}) => {
  return (
    <span className="animate-pulse">
      <svg
        className={cn('animate-spin', {
          'h-5 w-5': variant === 'default',
          'h-10 w-10': variant === 'large',
        })}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="8"
          stroke="currentColor"
          strokeWidth={strokeWidth === 'thick' ? 4 : 2}
        ></circle>
        <path
          className="opacity-75"
          stroke="currentColor"
          strokeWidth={strokeWidth === 'thick' ? 4 : 2}
          strokeLinecap="round"
          d="M4 12a8 8 0 018-8"
        ></path>
      </svg>
    </span>
  )
}
