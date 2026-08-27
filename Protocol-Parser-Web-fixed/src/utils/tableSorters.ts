type DateValue = string | number | Date | null | undefined;

const toTimestamp = (value: DateValue): number | null => {
  if (value === null || value === undefined || value === "") return null;
  if (value instanceof Date) return value.getTime();
  if (typeof value === "number") return Number.isFinite(value) ? value : null;
  const timestamp = Date.parse(value.trim().replace(" ", "T"));
  return Number.isNaN(timestamp) ? null : timestamp;
};

export const compareDateValues = (left: DateValue, right: DateValue) => {
  const leftTime = toTimestamp(left);
  const rightTime = toTimestamp(right);
  if (leftTime === null && rightTime === null) return 0;
  if (leftTime === null) return 1;
  if (rightTime === null) return -1;
  return leftTime - rightTime;
};

export const dateColumnSorter = <T>(field: keyof T) => ({
  sorter: (left: T, right: T) => compareDateValues(left[field] as DateValue, right[field] as DateValue),
  sortDirections: ["ascend", "descend", "ascend"] as Array<"ascend" | "descend">,
});

const defaultDateFields = new Set([
  "time", "createdAt", "updatedAt", "recordDate", "handledAt", "boundAt", "delegatedAt",
  "registrationDate", "inboundDate", "insuranceExpiry", "inspectionExpiry",
  "driverLicenseIssueDate", "driverLicenseExpiry", "drivingLicenseIssueDate",
  "serviceStartTime", "serviceEndTime",
]);

export const withDateColumnSorters = <T,>(columns: Array<Record<string, any>>) => columns.map((column) => {
  const field = String(column.dataIndex ?? "");
  const dateTitle = typeof column.title === "string" && /(时间|日期|到期)/.test(column.title);
  return field && (defaultDateFields.has(field) || dateTitle)
    ? { ...column, ...dateColumnSorter<T>(field as keyof T) }
    : column;
});
