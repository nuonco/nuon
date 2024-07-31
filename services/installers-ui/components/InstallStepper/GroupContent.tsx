import { Accordion, AccordionHeader, AccordionBody } from "../Accordion";
import { Markdown } from "@/components/Markdown";
import { InputFields, Card } from "@/components";

export const GroupContent = ({
  group = { display_name: "", description: "" },
  idx = 0,
  activeStep = 0,
  setActiveStep = (idx = 0) => {},
  searchParams = {},
  install_input_values = {},
}) => (
  <Accordion key={idx} open={activeStep === idx + 1}>
    <AccordionHeader
      onClick={() => setActiveStep(idx + 1)}
      className={
        activeStep === idx + 1
          ? `px-4
              text-accordion-header-active-color
              dark:text-accordion-header-active-color-dark
              hover:!text-gray-500
              bg-accordion-header-active-background
              dark:bg-accordion-header-active-background-dark
          `
          : "px-4"
      }
    >
      {group.display_name}
    </AccordionHeader>
    <AccordionBody>
      <Card>
        <div className="px-4">
          <Markdown content={group.description}></Markdown>
        </div>
        <InputFields
          group={group}
          install_input_values={install_input_values}
          searchParams={searchParams}
        />
      </Card>
    </AccordionBody>
  </Accordion>
);
