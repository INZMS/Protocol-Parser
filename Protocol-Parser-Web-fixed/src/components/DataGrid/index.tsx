import { useEffect, useLayoutEffect, useMemo, useRef, useState, type Key, type ReactNode } from "react";
import { Button, Dropdown, message, Pagination, Space, Table } from "antd";
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

// Ant Design renders expanded tree rows in addition to their root records.
// Counting only `dataSource.length` makes a fully expanded menu tree look
// short to the layout engine, so it can grow below the viewport with no
// internal scrollbar.  Count the rendered tree instead.
const countTreeRows = (rows: unknown[]): number => rows.reduce<number>((count, row) => {
  const children = row && typeof row === "object" && Array.isArray((row as { children?: unknown[] }).children)
    ? (row as { children: unknown[] }).children
    : [];
  return count + 1 + countTreeRows(children);
}, 0);

// Ant Table receives tree roots and expands their descendants internally.  A
// mobile card list does not have that behaviour, so it must flatten the same
// tree itself instead of silently rendering only the root records.
const flattenVisibleTree = <T extends object>(rows: T[], getKey: (row: T) => Key, expandedKeys?: readonly Key[]): T[] => {
  const expanded = expandedKeys ? new Set(expandedKeys.map(String)) : null;
  return rows.flatMap(row => {
    const children = (row as { children?: T[] }).children;
    const showChildren = !expanded || expanded.has(String(getKey(row)));
    return [row, ...(children?.length && showChildren ? flattenVisibleTree(children, getKey, expandedKeys) : [])];
  });
};

export default function DataGrid<T extends object>({ title, preferenceKey, columns, dataSource, rowKey, loading, rowSelection, expandable, pagination, scroll, cardRender }: DataGridProps<T>) {
  const userId = useAuthStore(state => state.user?.id);
  const allKeys = useMemo(() => columns.map((column, index) => keyOf(column as Record<string, unknown>, index)), [columns]);
  const [visibleKeys, setVisibleKeys] = useState<string[] | null>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  const [viewportHeight, setViewportHeight] = useState(0);
  const [viewportWidth, setViewportWidth] = useState(0);
  const [internalSelectedKeys, setInternalSelectedKeys] = useState<Key[]>([]);

  useEffect(() => {
    let active = true;
    setVisibleKeys(null);
    void loadUserPreference<string[]>(preferenceKey).then(value => {
      if (active) setVisibleKeys(Array.isArray(value) ? value : null);
    }).catch(() => { if (active) setVisibleKeys(null); });
    return () => { active = false; };
  }, [preferenceKey, userId]);

  useLayoutEffect(() => {
    const element = viewportRef.current;
    if (!element) return;
    const update = () => {
      const rect = element.getBoundingClientRect();
      setViewportHeight(Math.floor(rect.height));
      setViewportWidth(Math.floor(rect.width));
    };
    update();
    const observer = new ResizeObserver(update);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  const selectedKeys = visibleKeys ?? allKeys;
  const visibleColumns = useMemo(() => {
    const selectedColumns = columns
      .filter((column, index) => selectedKeys.includes(keyOf(column as Record<string, unknown>, index)))
      .map(column => column.title === "操作"
        ? { ...column, className: `${column.className ?? ""} data-grid__operation-cell`.trim() }
        : column);
    return withDateColumnSorters<T>(selectedColumns as Array<Record<string, unknown>>) as ColumnsType<T>;
  }, [columns, selectedKeys]);
  const resolveRowKey = (record: T): Key => typeof rowKey === "function" ? rowKey(record) : record[rowKey] as Key;
  const mobileDataSource = useMemo(
    () => flattenVisibleTree(dataSource, resolveRowKey, expandable?.expandedRowKeys),
    [dataSource, expandable?.expandedRowKeys, resolveRowKey],
  );
  useEffect(() => {
    const available = new Set(mobileDataSource.map(resolveRowKey));
    setInternalSelectedKeys(current => {
      const next = current.filter(key => available.has(key));
      return next.length === current.length ? current : next;
    });
  }, [mobileDataSource, rowKey]);
  const updateColumns = (keys: Key[]) => {
    const next = keys.map(String);
    if (!next.length) {
      message.warning("至少保留一个列表字段");
      return;
    }
    setVisibleKeys(next);
    void saveUserPreference(preferenceKey, next).catch(() => message.error("字段显示设置保存失败"));
  };
  const paginationProps = pagination === false ? null : (() => {
    const { position: _position, ...props } = pagination ?? {};
    return props;
  })();
  // Keep ten rows naturally sized.  Only hand scrolling to Ant Design when the
  // data really exceeds the content viewport; this avoids both fake scrollbars
  // and the large blank area caused by a fixed scroll.y on short lists.
  const headerHeight = 42;
  const rowHeight = 44;
  // Reserve a bounded, scrollable tree viewport whenever a grid can expand.
  // A user can open nodes after first render even when defaultExpandAllRows is
  // false, so this must not depend on the initial expansion flag.
  // For a tree, use its complete data height rather than the current expanded
  // row count.  Collapsing a node then cannot make the observer alternately
  // add/remove scroll.y, which was the source of the visible table jitter.
  const stableRowCount = expandable ? countTreeRows(dataSource) : dataSource.length;
  const requiredTableHeight = headerHeight + stableRowCount * rowHeight;
  const needsBodyScroll = viewportHeight > 0 && requiredTableHeight > viewportHeight;
  // `scroll.y` is the body height, while the header is rendered above it.
  // Reserve both the header and the horizontal scrollbar so the final record
  // is never obscured by the bottom edge of the scroll container.
  const bodyScrollHeight = Math.max(160, viewportHeight - headerHeight - 10);
  // Do not force a horizontal scrollbar just because a page supplied an
  // optional x width. It appears only once the visible columns cannot fit.
  // Every list uses the same selection column.  Let the available viewport,
  // rather than a page-specific `scroll.x`, decide whether horizontal scrolling
  // is needed.  This keeps normal-width tables (for example menu management)
  // free of a redundant bottom scrollbar while preserving access to wide ones.
  const estimatedTableWidth = 48 + visibleColumns.reduce((total, column) => {
    const width = (column as { width?: number | string }).width;
    return total + (typeof width === "number" ? width : 140);
  }, 0);
  const needsHorizontalScroll = viewportWidth > 0 && estimatedTableWidth > viewportWidth;
  const baseScroll = needsHorizontalScroll
    ? { ...scroll, x: scroll?.x ?? "max-content" }
    : scroll?.y
      ? { ...scroll, x: undefined }
      : undefined;
  const resolvedScroll = needsBodyScroll
    ? { ...baseScroll, y: bodyScrollHeight }
    : baseScroll;
  const effectiveRowSelection: TableProps<T>["rowSelection"] = rowSelection ?? {
    selectedRowKeys: internalSelectedKeys,
    onChange: keys => setInternalSelectedKeys(keys),
  };
  const selectedCardKeys = effectiveRowSelection?.selectedRowKeys ?? internalSelectedKeys;
  const updateCardSelection = (record: T, checked: boolean) => {
    const key = resolveRowKey(record);
    const next = checked ? [...new Set([...selectedCardKeys, key])] : selectedCardKeys.filter(item => item !== key);
    effectiveRowSelection?.onChange?.(next, mobileDataSource.filter(item => next.includes(resolveRowKey(item))), { type: "single" });
  };

  // The mobile representation is derived from the same columns as the desktop
  // table. This prevents page-local summaries from silently dropping fields
  // whenever a list gains a new column.
  const mobileCardRender = (record: T) => {
    const fields = visibleColumns
      .filter(column => typeof column.title === "string" && typeof column.dataIndex === "string" && column.dataIndex !== "id")
      .map((column, index) => {
        const value = record[column.dataIndex as keyof T];
        const rendered = column.render ? column.render(value, record, index) : String(value ?? "—");
        return <div className="data-card-list__field" key={String(column.key ?? column.dataIndex)}><span>{column.title}</span><strong>{rendered ?? "—"}</strong></div>;
      });
    const actions = visibleColumns
      .filter(column => !column.dataIndex && typeof column.render === "function")
      .map((column, index) => <>{column.render?.(undefined, record, index)}</>);
    return <><div className="data-card-list__fields">{fields}</div>{actions.length > 0 && <Space wrap className="data-card-list__actions">{actions}</Space>}</>;
  };

  return <section className={`data-grid data-grid--responsive${needsHorizontalScroll ? " data-grid--horizontal-scroll" : ""}`}>
    <header className="data-grid__toolbar">
      <strong>{title}</strong>
      <Dropdown trigger={["click"]} menu={{ selectable: true, multiple: true, selectedKeys, items: columns.map((column, index) => ({ key: keyOf(column as Record<string, unknown>, index), label: labelOf(column as Record<string, unknown>) })), onSelect: ({ selectedKeys: keys }) => updateColumns(keys), onDeselect: ({ selectedKeys: keys }) => updateColumns(keys) }}>
        <Button type="text" icon={<SettingOutlined />}>字段设置</Button>
      </Dropdown>
    </header>
    <div ref={viewportRef} className={`data-grid__table${needsBodyScroll ? " data-grid__table--scroll" : ""}`}>
      <Table<T> className="pro-list-table" size="middle" tableLayout="fixed" sticky={needsBodyScroll} rowKey={rowKey as TableProps<T>["rowKey"]} loading={loading} dataSource={dataSource} columns={visibleColumns} rowSelection={effectiveRowSelection} expandable={expandable} pagination={false} scroll={resolvedScroll} />
    </div>
    <CardList items={mobileDataSource} rowKey={rowKey} renderItem={mobileCardRender} loading={loading} selectedKeys={selectedCardKeys} onSelect={updateCardSelection} />
    {paginationProps ? <footer className="data-grid__pagination"><Pagination {...paginationProps} /></footer> : null}
  </section>;
}
