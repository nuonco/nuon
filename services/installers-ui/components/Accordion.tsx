import React, { type FC } from "react";
import {
  Accordion as MTAccordion,
  AccordionHeader as MTAccordionHeader,
  AccordionBody as MTAccordionBody,
} from "@material-tailwind/react";

export const Accordion = ({ children, className = "", open, ...props }) => {
  return (
    <MTAccordion className={`${className}`} open={open} {...props}>
      {children}
    </MTAccordion>
  );
};

export const AccordionHeader = ({ children, className = "", ...props }) => {
  return (
    <MTAccordionHeader
      className={`text-black dark:text-white hover:!text-gray-500 ${className}`}
      {...props}
    >
      {children}
    </MTAccordionHeader>
  );
};

export const AccordionBody = ({ children, className = "", ...props }) => {
  return (
    <MTAccordionBody
      className={`text-black dark:text-white ${className}`}
      {...props}
    >
      {children}
    </MTAccordionBody>
  );
};
