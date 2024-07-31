import React, { type FC } from "react";

export const InputFields: FC<{
  group: Record<string, any>;
  searchParams?: Record<string, string>;
  install_input_values: Object;
}> = ({ group, searchParams = {}, install_input_values }) => {
  return (
    <fieldset key={group.id} name={group.name} className="p-4 w-full">
      {group.app_inputs.map((input) => (
        <label className="mb-2 flex flex-col flex-auto gap-2" key={input?.id}>
          <span className="text-sm font-medium">
            {input?.display_name || input?.name}
          </span>
          <input
            className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
            defaultValue={
              Object.keys(install_input_values).includes(input.name)
                ? install_input_values[input.name]
                : Object.hasOwn(searchParams, `input:${input.name}`)
                  ? searchParams?.[`input:${input.name}`]
                  : input?.default
            }
            name={`input:${input.name}`}
            required={input.required}
            type={input?.sensitive ? "password" : "text"}
          />
          {input?.description && (
            <span className="text-xs">{input?.description}</span>
          )}
        </label>
      ))}
    </fieldset>
  );
};
