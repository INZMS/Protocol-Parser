import type { ReactNode } from "react";

type ListPageShellProps = {
  query?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
};

/**
 * 生产列表页的唯一页面骨架。
 * 查询、业务操作和数据列表各自独立成区；页面本身不滚动，空间不足时
 * 只允许数据列表内部滚动，避免不同业务页面各写一套高度计算。
 */
export default function ListPageShell({ query, actions, children, className = "" }: ListPageShellProps) {
  return <section className={`list-page-shell ${className}`.trim()}>
    {query === undefined ? children : <>
      <section className="list-page-shell__query">{query}</section>
      {actions ? <section className="list-page-shell__actions">{actions}</section> : null}
      <section className="list-page-shell__content">{children}</section>
    </>}
  </section>;
}
