import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vitest";
import { Text } from "./Text";

describe("Text component", () => {
  test("renders with default props", () => {
    render(<Text>Hello World</Text>);
    
    const element = screen.getByText("Hello World");
    expect(element).toBeInTheDocument();
    expect(element.tagName).toBe("SPAN");
    expect(element).toHaveClass("font-sans", "text-sm", "leading-6", "tracking-[-0.2px]", "font-normal", "text-wrap");
  });

  test("renders as different semantic elements based on role", () => {
    const { rerender } = render(<Text role="paragraph">Paragraph text</Text>);
    expect(screen.getByText("Paragraph text").tagName).toBe("P");

    rerender(<Text role="heading" level={2}>Heading text</Text>);
    expect(screen.getByText("Heading text").tagName).toBe("H2");
    expect(screen.getByText("Heading text")).toHaveAttribute("aria-level", "2");
    expect(screen.getByText("Heading text")).toHaveAttribute("role", "heading");

    rerender(<Text role="code">Code text</Text>);
    expect(screen.getByText("Code text").tagName).toBe("CODE");

    rerender(<Text role="time">Time text</Text>);
    expect(screen.getByText("Time text").tagName).toBe("TIME");
  });

  test("applies correct font family classes", () => {
    const { rerender } = render(<Text family="sans">Sans text</Text>);
    expect(screen.getByText("Sans text")).toHaveClass("font-sans");

    rerender(<Text family="mono">Mono text</Text>);
    expect(screen.getByText("Mono text")).toHaveClass("font-mono");
  });

  test("applies correct variant classes", () => {
    const { rerender } = render(<Text variant="h1">H1 text</Text>);
    expect(screen.getByText("H1 text")).toHaveClass("text-[34px]", "leading-10", "tracking-[-0.8px]");

    rerender(<Text variant="h2">H2 text</Text>);
    expect(screen.getByText("H2 text")).toHaveClass("text-2xl", "leading-[30px]", "tracking-[-0.8px]");

    rerender(<Text variant="h3">H3 text</Text>);
    expect(screen.getByText("H3 text")).toHaveClass("text-lg", "leading-[27px]", "tracking-[-0.2px]");

    rerender(<Text variant="base">Base text</Text>);
    expect(screen.getByText("Base text")).toHaveClass("text-base", "leading-6", "tracking-[-0.2px]");

    rerender(<Text variant="body">Body text</Text>);
    expect(screen.getByText("Body text")).toHaveClass("text-sm", "leading-6", "tracking-[-0.2px]");

    rerender(<Text variant="subtext">Subtext</Text>);
    expect(screen.getByText("Subtext")).toHaveClass("text-xs", "leading-[17px]", "tracking-[-0.2px]");

    rerender(<Text variant="label">Label text</Text>);
    expect(screen.getByText("Label text")).toHaveClass("text-[11px]", "leading-[14px]", "tracking-[-0.2px]");
  });

  test("applies correct weight classes", () => {
    const { rerender } = render(<Text weight="normal">Normal text</Text>);
    expect(screen.getByText("Normal text")).toHaveClass("font-normal");

    rerender(<Text weight="strong">Strong text</Text>);
    expect(screen.getByText("Strong text")).toHaveClass("font-strong");

    rerender(<Text weight="stronger">Stronger text</Text>);
    expect(screen.getByText("Stronger text")).toHaveClass("font-stronger");
  });

  test("applies correct theme classes", () => {
    const { rerender } = render(<Text theme="default">Default text</Text>);
    expect(screen.getByText("Default text")).not.toHaveClass("text-cool-grey-600");

    rerender(<Text theme="neutral">Neutral text</Text>);
    expect(screen.getByText("Neutral text")).toHaveClass("text-cool-grey-600", "dark:text-white/70");

    rerender(<Text theme="info">Info text</Text>);
    expect(screen.getByText("Info text")).toHaveClass("text-blue-800", "dark:text-blue-600");

    rerender(<Text theme="warn">Warn text</Text>);
    expect(screen.getByText("Warn text")).toHaveClass("text-orange-800", "dark:text-orange-600");

    rerender(<Text theme="error">Error text</Text>);
    expect(screen.getByText("Error text")).toHaveClass("text-red-800", "dark:text-red-500");

    rerender(<Text theme="success">Success text</Text>);
    expect(screen.getByText("Success text")).toHaveClass("text-green-800", "dark:text-green-500");

    rerender(<Text theme="brand">Brand text</Text>);
    expect(screen.getByText("Brand text")).toHaveClass("text-primary-600", "dark:text-primary-500");
  });

  test("applies special tracking for headings with mono font", () => {
    const { rerender } = render(<Text family="mono" variant="h1">Mono H1</Text>);
    expect(screen.getByText("Mono H1")).toHaveClass("tracking-[-0.2px]");

    rerender(<Text family="mono" variant="h2">Mono H2</Text>);
    expect(screen.getByText("Mono H2")).toHaveClass("tracking-[-0.2px]");

    rerender(<Text family="mono" variant="h3">Mono H3</Text>);
    expect(screen.getByText("Mono H3")).toHaveClass("tracking-[-0.2px]");

    // Should not apply special tracking for non-heading variants
    rerender(<Text family="mono" variant="body">Mono body</Text>);
    const bodyElement = screen.getByText("Mono body");
    expect(bodyElement.className).not.toMatch(/tracking-\[-0\.2px\].*tracking-\[-0\.2px\]/);
  });

  test("forwards additional HTML attributes", () => {
    render(<Text data-testid="custom-text" title="Custom title">Text with attributes</Text>);
    
    const element = screen.getByTestId("custom-text");
    expect(element).toHaveAttribute("title", "Custom title");
  });

  test("merges custom className with component classes", () => {
    render(<Text className="custom-class">Custom styled text</Text>);
    
    const element = screen.getByText("Custom styled text");
    expect(element).toHaveClass("custom-class");
    expect(element).toHaveClass("font-sans"); // Should still have default classes
  });

  test("handles all heading levels", () => {
    const { rerender } = render(<Text role="heading" level={1}>H1</Text>);
    expect(screen.getByText("H1").tagName).toBe("H1");
    expect(screen.getByText("H1")).toHaveAttribute("aria-level", "1");

    rerender(<Text role="heading" level={3}>H3</Text>);
    expect(screen.getByText("H3").tagName).toBe("H3");
    expect(screen.getByText("H3")).toHaveAttribute("aria-level", "3");

    rerender(<Text role="heading" level={6}>H6</Text>);
    expect(screen.getByText("H6").tagName).toBe("H6");
    expect(screen.getByText("H6")).toHaveAttribute("aria-level", "6");
  });

  test("does not set aria-level without heading role", () => {
    render(<Text level={2}>Not a heading</Text>);
    
    const element = screen.getByText("Not a heading");
    expect(element.tagName).toBe("SPAN");
    expect(element).not.toHaveAttribute("aria-level");
    expect(element).not.toHaveAttribute("role");
  });

  test("combines multiple props correctly", () => {
    render(
      <Text 
        family="mono" 
        variant="h2" 
        weight="stronger" 
        theme="brand" 
        role="heading" 
        level={2}
        className="extra-class"
      >
        Complex text
      </Text>
    );
    
    const element = screen.getByText("Complex text");
    expect(element.tagName).toBe("H2");
    expect(element).toHaveClass(
      "font-mono", 
      "text-2xl", 
      "font-stronger", 
      "text-primary-600",
      "tracking-[-0.2px]", // Special mono heading tracking
      "extra-class"
    );
    expect(element).toHaveAttribute("aria-level", "2");
    expect(element).toHaveAttribute("role", "heading");
  });
});