export const ErrorMessage = ({ message }: { message: string }) => (
  <div className="rounded-md bg-red-50 p-4">
    <div className="flex">
      <div className="text-sm text-red-700">{message}</div>
    </div>
  </div>
)
