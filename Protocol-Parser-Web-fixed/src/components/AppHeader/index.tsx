import { useState } from "react";
import { CodeOutlined, QuestionCircleOutlined } from "@ant-design/icons";
import { Drawer } from "antd";

export default function Header() {
    const [helpOpen, setHelpOpen] = useState(false);

    return (
        <>
        <div
            style={{
                height: 64,
                background: "#fff",
                borderBottom: "1px solid #e5e7eb",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "0 clamp(14px, 4vw, 28px)",
                gap: 12
            }}
        >
            {/* 左侧 */}
            <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
                <div
                    style={{
                        flex: "0 0 auto",
                        width: 34,
                        height: 34,
                        borderRadius: 8,
                        border: "1px solid #d9d9d9",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        color: "#1677ff",
                        fontSize: 16
                    }}
                >
                    <CodeOutlined />
                </div>

                <div style={{ minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, lineHeight: "20px", whiteSpace: "nowrap" }}>
                        协议解析工具
                    </div>
                    <div style={{ fontSize: 11, color: "#8c8c8c", lineHeight: "16px" }}>
                        Protocol Parser Tool
                    </div>
                </div>
            </div>

            {/* 右侧 */}
            <div style={{ display: "flex", alignItems: "center", gap: 20, flex: "0 0 auto", fontSize: 13 }}>
                <button className="header-help-button" type="button" onClick={() => setHelpOpen(true)}>
                    <QuestionCircleOutlined />
                    <span>使用说明</span>
                </button>

                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <div
                        style={{
                            width: 28,
                            height: 28,
                            borderRadius: "50%",
                            background: "#e6f4ff",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                            fontSize: 12
                        }}
                    >
                        A
                    </div>
                    admin
                </div>
            </div>
        </div>

        <Drawer
            title="协议解析工具 · 使用说明"
            placement="right"
            width="min(600px, 92vw)"
            className="help-drawer"
            open={helpOpen}
            onClose={() => setHelpOpen(false)}
        >
            <div className="help-content">
                <section className="help-intro">
                    <h3>工具用途</h3>
                    <p>
                        本工具用于将设备上报的十六进制协议报文转换为可阅读、可核对的结构化数据，
                        帮助研发、测试、实施和技术支持人员快速定位报文内容及字段异常。
                    </p>
                </section>

                <section>
                    <h3>为什么使用它</h3>
                    <ul>
                        <li>减少人工按字节查协议文档和换算字段的时间。</li>
                        <li>同时保留原始 HEX、字段偏移、字段长度和最终解析值，便于逐项核对。</li>
                        <li>将时间、经纬度、伪 IP、基站、WiFi、工作模式等数据转换为业务可读结果。</li>
                        <li>保留解析历史，方便问题复现、结果复制及前后报文对比。</li>
                    </ul>
                </section>

                <section>
                    <h3>使用步骤</h3>
                    <ol>
                        <li>选择报文所属协议；当前提供 2929 协议。</li>
                        <li>输入完整 HEX 报文，空格、换行和制表符会自动忽略；也可点击报文示例填充。</li>
                        <li>点击“解析报文”，在表格视图查看字段级结果，在 JSON 视图查看完整业务数据。</li>
                        <li>解析成功后记录会保存到历史列表，可查看详情、复制完整结果或删除单条记录。</li>
                    </ol>
                </section>

                <section>
                    <h3>当前解析范围</h3>
                    <p>
                        当前支持 2929 和 JT-808。2929重点解析通用头和位置上报（0x80），包括定位时间（北京时间）、经纬度换算、
                        速度、方向、状态、基站、WiFi 热点、SIM ICCID、工作/上报模式及下次上报时间等字段；
                        中心确认（0x21）和申请设置参数（0xD8）也可解析对应字段。JT-808支持转义、XOR校验、通用消息头、
                        通用应答、心跳、注册/鉴权报文内容、终端参数、位置及批量位置、位置附加项和文本消息解析。
                    </p>
                </section>

                <section>
                    <h3>记录与数据说明</h3>
                    <p>
                        解析历史存储在 MySQL 中。同一协议下完全相同的原始报文会复用原记录，避免重复落库。
                        本工具只负责报文解析与记录查看，不包含设备注册、鉴权、在线管理、指令下发或 TCP 长连接功能。
                    </p>
                </section>

                <div className="help-note">
                    提示：解析结果用于辅助排查。遇到字段长度不足、校验失败或设备厂商存在扩展定义时，
                    应结合对应版本的协议文档和设备实际配置进行确认。
                </div>
            </div>
        </Drawer>
        </>
    );
}
