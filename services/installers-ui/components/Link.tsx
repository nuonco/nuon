import React, { type FC } from "react";
import NextLink from "next/link";

export const Link: FC<React.ComponentProps<typeof NextLink>> = ({
  className = "",
  children,
  ...props
}) => {
  return (
    <NextLink
      className={`text-primary-500 hover:text-primary-700 focus:text-primary-700 ${className}`}
      {...props}
    >
      {children}
    </NextLink>
  );
};
