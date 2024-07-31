import React, { type FC } from "react";
import { Stepper as MTStepper, Step as MTStep } from "@material-tailwind/react";

export const Stepper = ({ children, ...props }) => {
  return (
    <MTStepper
      lineClassName="bg-black dark:bg-white"
      activeLineClassName="!bg-primary-500"
      {...props}
    >
      {children}
    </MTStepper>
  );
};

export const Step = ({ children, ...props }) => {
  // NOTE: the classNames are not abstracted in here
  return <MTStep {...props}>{children}</MTStep>;
};
