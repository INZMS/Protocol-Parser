import type { Key, ReactNode } from "react";
import { Empty, Skeleton } from "antd";

type CardListProps<T> = {
  items: T[];
  rowKey: keyof T | ((item: T) => Key);
  renderItem: (item: T) => ReactNode;
  loading?: boolean;
};

export default function CardList<T>({ items, rowKey, renderItem, loading = false }: CardListProps<T>) {
  if (loading) return <div className="data-card-list"><Skeleton active /></div>;
  if (!items.length) return <div className="data-card-list"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /></div>;
  const getKey = (item: T) => typeof rowKey === "function" ? rowKey(item) : String(item[rowKey]);
  return <div className="data-card-list">{items.map(item => <article className="data-card-list__item" key={getKey(item)}>{renderItem(item)}</article>)}</div>;
}
