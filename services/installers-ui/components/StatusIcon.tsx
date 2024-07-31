import React, { type FC } from "react";

export const StatusIcon: FC<{ status: string; className?: string }> = ({
  status,
  className,
}) => {
  let clr = "yellow-500";
  if (status === "active") {
    clr = "green-600";
  } else if (status === "error") {
    clr = "red-500";
  } else if (status === "") {
    clr = "slate-300";
  }

  return <span className={`text-${clr} ${className}`}>⏺</span>;
};
