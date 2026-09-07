import type { Key, ReactNode } from "react";
import { Checkbox, Empty, Skeleton } from "antd";

type CardListProps<T> = {
  items: T[];
  rowKey: keyof T | ((item: T) => Key);
  renderItem: (item: T) => ReactNode;
  loading?: boolean;
  selectedKeys?: Key[];
  onSelect?: (item: T, checked: boolean) => void;
};

export default function CardList<T>({ items, rowKey, renderItem, loading = false, selectedKeys, onSelect }: CardListProps<T>) {
  if (loading) return <div className="data-card-list"><Skeleton active /></div>;
  if (!items.length) return <div className="data-card-list"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /></div>;
  const getKey = (item: T): Key => typeof rowKey === "function" ? rowKey(item) : item[rowKey] as Key;
  return <div className="data-card-list">{items.map(item => {
    const key = getKey(item);
    return <article className="data-card-list__item" key={key}>
      {onSelect && <Checkbox className="data-card-list__selection" checked={selectedKeys?.includes(key)} onChange={event => onSelect(item, event.target.checked)} />}
      {renderItem(item)}
    </article>;
  })}</div>;
}
