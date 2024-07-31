"use client";

import React, { type FC, useState } from "react";
import ReactSelect from "react-tailwindcss-select";

interface Option {
  value: string;
  label: string;
  disabled?: boolean;
  isSelected?: boolean;
}

const classNames = {
  menuButton: ({ isDisabled }) =>
    `flex text-sm text-gray-500 border border-gray-300 rounded shadow-sm transition-all duration-300 focus:outline-none ${
      isDisabled
        ? "bg-gray-200"
        : "bg-white hover:border-gray-400 focus:border-primary-500 focus:ring focus:ring-primary-500/20"
    }`,
  listItem: ({ isSelected }) =>
    `block transition duration-200 px-2 py-2 cursor-pointer select-none truncate rounded ${
      isSelected
        ? `text-white bg-primary-500`
        : `text-gray-500 hover:bg-primary-200 hover:text-primary-600`
    }`,
};

export const Select: FC<{
  defaultValue?: string;
  name: string;
  options: Option[];
  required?: boolean;
}> = ({ defaultValue = "", name, options, required = false }) => {
  const [value, setValue] = useState<Option>(
    options.find((o) => o.value === defaultValue) || options[0],
  );

  return (
    <>
      <input
        type="hidden"
        name={name}
        value={value.value}
        required={required}
      />
      <ReactSelect
        value={value}
        onChange={(e) => {
          setValue(e as any);
        }}
        options={options as any}
        /* we're not using primaryColor, but if it's not set the build fails due to TS errors */
        primaryColor="blue"
        classNames={classNames}
      />
    </>
  );
};
