export const LoadingSpinner = ({ className = '' }: { className?: string }) => (
  <div className={`flex items-center justify-center py-12 ${className}`}>
    <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-200 border-t-primary-600" />
  </div>
)
