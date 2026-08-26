import { Button, Result } from "antd";
import type { ReactNode } from "react";

export type ExceptionCode=403|404|500;
const content:Record<ExceptionCode,{title:string;description:string}>={
  403:{title:"403",description:"抱歉，您没有权限访问当前页面。"},
  404:{title:"404",description:"抱歉，您访问的页面不存在。"},
  500:{title:"500",description:"抱歉，系统运行时出现异常，请稍后重试。"},
};

export default function ExceptionPage({code,extra}:{code:ExceptionCode;extra?:ReactNode}){
  const item=content[code];
  return <div className="exception-page"><Result status={String(code) as "403"|"404"|"500"} title={item.title} subTitle={item.description} extra={extra??<Button type="primary" onClick={()=>window.location.assign("/")}>返回首页</Button>}/></div>;
}
