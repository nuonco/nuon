export const ErrorMessage = ({ message }: { message: string }) => (
  <div className="rounded-md bg-red-50 p-4 dark:bg-red-900/30">
    <div className="flex">
      <div className="text-sm text-red-700 dark:text-red-300">{message}</div>
    </div>
  </div>
)
