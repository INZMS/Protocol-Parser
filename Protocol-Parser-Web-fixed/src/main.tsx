import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

import App from './App'
import AppErrorBoundary from './components/AppErrorBoundary'
import { configureHttpClient } from './api/client'

import './index.css'

dayjs.locale('zh-cn')
configureHttpClient()

const validateMessages={
  required:'请输入${label}',
  types:{email:'请输入有效的${label}',number:'${label}必须是有效数字'},
  string:{min:'${label}不能少于${min}个字符',max:'${label}不能超过${max}个字符'},
}


ReactDOM.createRoot(
  document.getElementById('root')!
)
.render(

  <React.StrictMode>

    <ConfigProvider locale={zhCN} form={{validateMessages}} theme={{token:{colorPrimary:'#1677ff',borderRadius:6,colorBgLayout:'#f5f7fa'}}}>
      <AppErrorBoundary><App /></AppErrorBoundary>
    </ConfigProvider>

  </React.StrictMode>

)
