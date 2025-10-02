import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createFileDownload, downloadFileOnClick } from "./file-download";

Object.defineProperty(global, "URL", {
  value: {
    createObjectURL: vi.fn(),
    revokeObjectURL: vi.fn(),
  },
  writable: true,
});

describe("file-download", () => {
  describe("createFileDownload", () => {
    let mockLink: HTMLAnchorElement;
    let mockCreateElement: ReturnType<typeof vi.fn>;
    let mockAppendChild: ReturnType<typeof vi.fn>;
    let mockRemoveChild: ReturnType<typeof vi.fn>;
    let mockClick: ReturnType<typeof vi.fn>;
    let mockSetAttribute: ReturnType<typeof vi.fn>;

    beforeEach(() => {
      vi.clearAllMocks();

      mockClick = vi.fn();
      mockSetAttribute = vi.fn();
      mockLink = {
        href: "",
        click: mockClick,
        setAttribute: mockSetAttribute,
      } as unknown as HTMLAnchorElement;

      mockCreateElement = vi.fn().mockReturnValue(mockLink);
      mockAppendChild = vi.fn();
      mockRemoveChild = vi.fn();

      Object.defineProperty(document, "createElement", {
        value: mockCreateElement,
        writable: true,
      });

      Object.defineProperty(document.body, "appendChild", {
        value: mockAppendChild,
        writable: true,
      });

      Object.defineProperty(document.body, "removeChild", {
        value: mockRemoveChild,
        writable: true,
      });

      (global.URL.createObjectURL as ReturnType<typeof vi.fn>).mockReturnValue(
        "blob:mock-url",
      );
    });

    test("should create download for string content with default mime type", () => {
      const content = "Hello, world!";
      const filename = "test.txt";

      createFileDownload(content, filename);

      expect(mockCreateElement).toHaveBeenCalledWith("a");
      expect(mockLink.href).toBe("blob:mock-url");
      expect(mockSetAttribute).toHaveBeenCalledWith("download", filename);
      expect(mockAppendChild).toHaveBeenCalledWith(mockLink);
      expect(mockClick).toHaveBeenCalled();
      expect(mockRemoveChild).toHaveBeenCalledWith(mockLink);
      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.any(Blob),
      );
      expect(global.URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");
    });

    test("should create download for string content with custom mime type", () => {
      const content = '{"key": "value"}';
      const filename = "data.json";
      const mimeType = "application/json";

      createFileDownload(content, filename, mimeType);

      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.objectContaining({
          type: mimeType,
        }),
      );
    });

    test("should handle Blob content directly", () => {
      const blob = new Blob(["test content"], { type: "text/plain" });
      const filename = "test.txt";

      createFileDownload(blob, filename);

      expect(global.URL.createObjectURL).toHaveBeenCalledWith(blob);
      expect(mockClick).toHaveBeenCalled();
    });

    test("should handle ArrayBuffer content", () => {
      const buffer = new ArrayBuffer(8);
      const view = new Uint8Array(buffer);
      view.set([72, 101, 108, 108, 111]);
      const filename = "binary.bin";
      const mimeType = "application/octet-stream";

      createFileDownload(buffer, filename, mimeType);

      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.objectContaining({
          type: mimeType,
        }),
      );
      expect(mockClick).toHaveBeenCalled();
    });

    test("should use default mime type when not specified", () => {
      const content = "test content";
      const filename = "test.txt";

      createFileDownload(content, filename);

      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.objectContaining({
          type: "text/plain",
        }),
      );
    });

    test("should clean up resources after download", () => {
      const content = "test content";
      const filename = "test.txt";

      createFileDownload(content, filename);

      expect(mockRemoveChild).toHaveBeenCalledWith(mockLink);
      expect(global.URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");
    });

    test("should handle special characters in filename", () => {
      const content = "test content";
      const filename = "test file (1).txt";

      createFileDownload(content, filename);

      expect(mockSetAttribute).toHaveBeenCalledWith("download", filename);
    });

    test("should handle empty string content", () => {
      const content = "";
      const filename = "empty.txt";

      createFileDownload(content, filename);

      expect(mockClick).toHaveBeenCalled();
      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.any(Blob),
      );
    });
  });

  describe("downloadFileOnClick", () => {
    let mockLink: HTMLAnchorElement;
    let mockCreateElement: ReturnType<typeof vi.fn>;
    let mockAppendChild: ReturnType<typeof vi.fn>;
    let mockRemoveChild: ReturnType<typeof vi.fn>;
    let mockClick: ReturnType<typeof vi.fn>;
    let mockCallback: ReturnType<typeof vi.fn>;

    beforeEach(() => {
      vi.clearAllMocks();
      vi.useFakeTimers();

      mockClick = vi.fn();
      mockCallback = vi.fn();
      mockLink = {
        href: "",
        download: "",
        click: mockClick,
      } as unknown as HTMLAnchorElement;

      mockCreateElement = vi.fn().mockReturnValue(mockLink);
      mockAppendChild = vi.fn();
      mockRemoveChild = vi.fn();

      Object.defineProperty(document, "createElement", {
        value: mockCreateElement,
        writable: true,
      });

      Object.defineProperty(document.body, "appendChild", {
        value: mockAppendChild,
        writable: true,
      });

      Object.defineProperty(document.body, "removeChild", {
        value: mockRemoveChild,
        writable: true,
      });

      (global.URL.createObjectURL as ReturnType<typeof vi.fn>).mockReturnValue(
        "blob:mock-url",
      );
    });

    test("should download file with default parameters", () => {
      const content = "test content";
      const filename = "test.txt";

      downloadFileOnClick({ content, filename });

      expect(mockCreateElement).toHaveBeenCalledWith("a");
      expect(mockLink.href).toBe("blob:mock-url");
      expect(mockLink.download).toBe("test.txt");
      expect(mockAppendChild).toHaveBeenCalledWith(mockLink);
      expect(mockClick).toHaveBeenCalled();
      expect(mockRemoveChild).toHaveBeenCalledWith(mockLink);
    });

    test("should use custom mime type when provided", () => {
      const content = '{"key": "value"}';
      const filename = "data.json";
      const mimeType = "application/json";

      downloadFileOnClick({ content, filename, mimeType });

      expect(global.URL.createObjectURL).toHaveBeenCalledWith(
        expect.objectContaining({
          type: mimeType,
        }),
      );
    });

    test("should use default filename with fileType when filename is not provided", () => {
      const content = "test content";
      const fileType = "yaml";

      downloadFileOnClick({ content, filename: "", fileType });

      expect(mockLink.download).toBe("download.yaml");
    });

    test("should clean filename by removing quotes and underscores", () => {
      const content = "test content";
      const filename = '"_test_file_"';

      downloadFileOnClick({ content, filename });

      expect(mockLink.download).toBe("test_file");
    });

    test("should execute callback when provided", () => {
      const content = "test content";
      const filename = "test.txt";

      downloadFileOnClick({ content, filename, callback: mockCallback });

      expect(mockCallback).toHaveBeenCalled();
    });

    test("should revoke object URL after timeout", () => {
      const content = "test content";
      const filename = "test.txt";

      downloadFileOnClick({ content, filename });

      expect(global.URL.revokeObjectURL).not.toHaveBeenCalled();

      vi.advanceTimersByTime(10000);

      expect(global.URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");
    });

    test("should handle empty filename with default fileType", () => {
      const content = "test content";
      const filename = "";

      downloadFileOnClick({ content, filename });

      expect(mockLink.download).toBe("download.toml");
    });

    test("should trim whitespace from cleaned filename", () => {
      const content = "test content";
      const filename = "  test.txt  ";

      downloadFileOnClick({ content, filename });

      expect(mockLink.download).toBe("test.txt");
    });

    afterEach(() => {
      vi.useRealTimers();
    });
  });
});
