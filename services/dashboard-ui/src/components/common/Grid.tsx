import React, { type FC } from 'react'

export const Grid: FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="grid auto-rows-auto gap-6 grid-cols-auto w-full">
      {children}
    </div>
  )
}
