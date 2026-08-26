import { Component, type ErrorInfo, type ReactNode } from "react";
import ExceptionPage from "../ExceptionPage";

export default class AppErrorBoundary extends Component<{children:ReactNode},{failed:boolean}>{
  state={failed:false};
  static getDerivedStateFromError(){return {failed:true}}
  componentDidCatch(error:Error,info:ErrorInfo){console.error("页面渲染异常",error,info)}
  render(){return this.state.failed?<ExceptionPage code={500}/>:this.props.children}
}
