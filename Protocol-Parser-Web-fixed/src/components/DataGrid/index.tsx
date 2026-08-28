import { useEffect, useMemo, useState, type Key, type ReactNode } from "react";
import { Button, Dropdown, message, Table } from "antd";
import { SettingOutlined } from "@ant-design/icons";
import type { ColumnsType, TablePaginationConfig, TableProps } from "antd/es/table";
import { useAuthStore } from "../../store/auth";
import { loadUserPreference, saveUserPreference } from "../../store/settings";
import { withDateColumnSorters } from "../../utils/tableSorters";
import CardList from "../CardList";

export type DataGridProps<T extends object> = {
  title: string;
  preferenceKey: string;
  columns: ColumnsType<T>;
  dataSource: T[];
  rowKey: keyof T | ((record: T) => Key);
  loading?: boolean;
  rowSelection?: TableProps<T>["rowSelection"];
  expandable?: TableProps<T>["expandable"];
  pagination?: TablePaginationConfig | false;
  scroll?: TableProps<T>["scroll"];
  cardRender?: (record: T) => ReactNode;
};

const keyOf = (column: Record<string, unknown>, index: number) => String(column.key ?? column.dataIndex ?? `column-${index}`);
const labelOf = (column: Record<string, unknown>) => typeof column.title === "string" ? column.title : "未命名字段";

export default function DataGrid<T extends object>({ title, preferenceKey, columns, dataSource, rowKey, loading, rowSelection, expandable, pagination, scroll, cardRender }: DataGridProps<T>) {
  const userId = useAuthStore(state => state.user?.id);
  const allKeys = useMemo(() => columns.map((column, index) => keyOf(column as Record<string, unknown>, index)), [columns]);
  const [visibleKeys, setVisibleKeys] = useState<string[] | null>(null);

  useEffect(() => {
    let active = true;
    setVisibleKeys(null);
    void loadUserPreference<string[]>(preferenceKey).then(value => {
      if (active) setVisibleKeys(Array.isArray(value) ? value : null);
    }).catch(() => { if (active) setVisibleKeys(null); });
    return () => { active = false; };
  }, [preferenceKey, userId]);

  const selectedKeys = visibleKeys ?? allKeys;
  const visibleColumns = useMemo(() => withDateColumnSorters<T>(columns.filter((column, index) => selectedKeys.includes(keyOf(column as Record<string, unknown>, index))) as Array<Record<string, unknown>>) as ColumnsType<T>, [columns, selectedKeys]);
  const updateColumns = (keys: Key[]) => {
    const next = keys.map(String);
    if (!next.length) {
      message.warning("至少保留一个列表字段");
      return;
    }
    setVisibleKeys(next);
    void saveUserPreference(preferenceKey, next).catch(() => message.error("字段显示设置保存失败"));
  };

  return <section className={`data-grid${cardRender ? " data-grid--responsive" : ""}`}>
    <header className="data-grid__toolbar">
      <strong>{title}</strong>
      <Dropdown trigger={["click"]} menu={{ selectable: true, multiple: true, selectedKeys, items: columns.map((column, index) => ({ key: keyOf(column as Record<string, unknown>, index), label: labelOf(column as Record<string, unknown>) })), onSelect: ({ selectedKeys: keys }) => updateColumns(keys), onDeselect: ({ selectedKeys: keys }) => updateColumns(keys) }}>
        <Button type="text" icon={<SettingOutlined />}>字段设置</Button>
      </Dropdown>
    </header>
    <div className="data-grid__table">
      <Table<T> className="pro-list-table" size="middle" tableLayout="fixed" rowKey={rowKey as TableProps<T>["rowKey"]} loading={loading} dataSource={dataSource} columns={visibleColumns} rowSelection={rowSelection} expandable={expandable} pagination={pagination} scroll={scroll} />
    </div>
    {cardRender && <CardList items={dataSource} rowKey={rowKey} renderItem={cardRender} loading={loading} />}
  </section>;
}
