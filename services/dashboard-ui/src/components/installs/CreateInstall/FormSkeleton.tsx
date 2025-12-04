import { Skeleton } from '@/components/common/Skeleton'

export const FormSkeleton = () => {
  return (
    <div className="flex flex-col gap-8 max-w-3xl">
      {/* Install name section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
        <span className="flex flex-col gap-1">
          <Skeleton width="100px" height="16px" />
          <Skeleton width="160px" height="14px" />
        </span>
        <Skeleton width="100%" height="40px" />
      </div>

      {/* AWS Settings */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <Skeleton width="140px" height="24px" />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="130px" height="16px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      {/* First input group */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="220px" height="24px" />
          <Skeleton width="280px" height="16px" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="80px" height="16px" />
            <Skeleton width="200px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      {/* Second input group */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="180px" height="24px" />
          <Skeleton width="140px" height="16px" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="90px" height="16px" />
            <Skeleton width="160px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="110px" height="16px" />
            <Skeleton width="240px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="85px" height="16px" />
            <Skeleton width="300px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>
    </div>
  )
}