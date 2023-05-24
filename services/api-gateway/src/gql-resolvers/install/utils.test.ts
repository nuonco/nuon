import { formatInstall, getInstallSettings } from "./utils";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  year: 1999,
};

const mockInstall = {
  awsSettings: {
    region: 1,
    role: "test:role",
  },
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

test("getInstallSettings should take a raw grpc install & format it with the correct settings options", () => {
  const spec = getInstallSettings(mockInstall);

  expect(spec).toEqual({
    __typename: "AWSSettings",
    region: "US_EAST_1",
    role: "test:role",
  });
});

test("formatInstall should return a GQL install with the correct date format & settings field", () => {
  const spec = formatInstall(mockInstall);

  expect(spec).toEqual({
    createdAt: "1999-12-31T08:15:30.000+00:00",
    id: "test-id",
    name: "test-node",
    settings: {
      __typename: "AWSSettings",
      region: "US_EAST_1",
      role: "test:role",
    },
    updatedAt: "1999-12-31T08:15:30.000+00:00",
  });
});
