import React, { type FC } from "react";

export const Card: FC<{ children?: React.ReactNode; className?: string }> = ({
  children,
  className = "",
}) => {
  return (
    <div
      className={`
        flex flex-col gap-2  items-start
        bg-card-background dark:bg-card-background-dark
        border-card-border-color dark:border-card-border-color-dark
        border-card-border-width
        rounded-card-radius ${className}`}
    >
      {children}
    </div>
  );
};
