import type { ReactNode } from "react";
import type { ColumnType } from "antd/es/table";
import { Tag } from "antd";
import { dateColumnSorter } from "./tableSorters";

export const textColumn = <T,>(title: string, dataIndex: keyof T, options: ColumnType<T> = {}): ColumnType<T> => ({ title, dataIndex: String(dataIndex), key: String(dataIndex), ellipsis: true, ...options });

export const dateColumn = <T,>(title: string, dataIndex: keyof T, options: ColumnType<T> = {}): ColumnType<T> => ({ ...textColumn<T>(title, dataIndex, options), ...dateColumnSorter<T>(dataIndex) });

export const statusColumn = <T,>(title: string, dataIndex: keyof T, labels: Record<string, { label: ReactNode; color?: string }>, options: ColumnType<T> = {}): ColumnType<T> => ({
  ...textColumn<T>(title, dataIndex, options),
  render: value => {
    const item = labels[String(value)];
    return item ? <Tag color={item.color}>{item.label}</Tag> : String(value ?? "—");
  },
});
