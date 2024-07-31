import { Accordion, AccordionHeader, AccordionBody } from "../Accordion";
import { Card } from "@/components";

export const CompanyContent = ({
  open = false,
  onClick = () => {},
  searchParams = { name: "" },
  name = null,
}) => (
  <Accordion open={open}>
    <AccordionHeader
      onClick={onClick}
      className={
        open
          ? `px-4
          text-accordion-header-active-color
          dark:text-accordion-header-active-color-dark
          hover:!text-gray-500
          bg-accordion-header-active-background
          dark:bg-accordion-header-active-background-dark`
          : "px-4"
      }
    >
      Company Info
    </AccordionHeader>
    <AccordionBody className="text-black dark:text-white">
      <Card>
        <fieldset className="p-4 w-full">
          <label className="flex flex-col flex-auto gap-2">
            <span className="text-sm font-medium">Name</span>
            <input
              className="border bg-inherit rounded px-4 py-1.5 shadow-inner"
              defaultValue={
                name
                  ? name
                  : Object.hasOwn(searchParams, "name")
                    ? searchParams.name
                    : ""
              }
              name="name"
              type="text"
              required
            />
          </label>
        </fieldset>
      </Card>
    </AccordionBody>
  </Accordion>
);
