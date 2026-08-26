import { Card, Row, Col, Tabs, Button, Empty,message } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import {useState} from "react";

import ParseTable from "../../components/ParseTable";
import { useParserStore } from "../../store/parser";
import { canAccess, useAuthStore } from "../../store/auth";

function StatBlock({ label, value, color }: { label: string; value: string | number; color?: string }) {
    return (
        <div className="stat-block">
            <div className="stat-label">{label}</div>
            <div className="stat-value" style={{ color }}>{value}</div>
        </div>
    );
}

export default function ResultPanel() {
    const { result } = useParserStore();
    const currentUser = useAuthStore(state => state.user);
    const [activeTab, setActiveTab] = useState("table");

    if (!result) {
        return (
            <Card title="解析结果" size="small">
                <div style={{ padding: "40px 0" }}>
                    <Empty description="暂无解析结果，请输入HEX报文后点击解析" />
                </div>
            </Card>
        );
    }

    return (
        <Card title="解析结果" size="small">
            {/* 基础信息：用普通文本块展示，避免 AntD Statistic 把 "0200" 这类
                字符串当数字加千分位，渲染成 "0,200" 的问题 */}
            <Row gutter={16}>
                <Col xs={12} sm={6}>
                    <StatBlock label="协议" value={result.protocol} />
                </Col>
                <Col xs={12} sm={6}>
                    <StatBlock label="消息ID" value={result.messageId} color="#389e0d" />
                </Col>
                <Col xs={12} sm={6}>
                    <StatBlock label="消息名称" value={result.messageName} color="#d46b08" />
                </Col>
                <Col xs={12} sm={6}>
                    <StatBlock label="报文长度" value={`${result.length} Bytes`} color="#722ed1" />
                </Col>
            </Row>

            {/* Tabs区域：表格视图 / JSON视图。
                "复制结果" 本质上是表格/JSON 两个视图共用的功能按钮，
                所以放进 Tabs 的 tabBarExtraContent，与视图切换同一行，
                而不是单独占一行悬在上面、把中间撑出一大块空白。 */}
            <Tabs
                className="result-tabs"
                style={{ marginTop: 12 }}
                activeKey={activeTab}
                onChange={(key) => setActiveTab(key)}
                tabBarExtraContent={{
                    right: canAccess(currentUser, "workbench:parser:copy") ? (
                        <Button
                            size="small"
                            type="primary"
                            icon={<CopyOutlined />}
                            onClick={() => {
                                if (activeTab === "table") {
                                    navigator.clipboard.writeText(
                                        JSON.stringify(result.fields, null, 2)
                                    );
                                    message.success("表格数据复制成功");
                                } else  {
                                    navigator.clipboard.writeText(
                                        JSON.stringify(result.data, null, 2)
                                    );
                                    message.success("JSON数据复制成功");
                                }
                            }}
                        >
                            复制结果
                        </Button>
                    ) : null
                }}
                items={[
                    {
                        key: "table",
                        label: "表格视图",
                        children: (
                            <>
                                <div className="result-panel-scroll">
                                    <ParseTable />
                                </div>
                                <div style={{ marginTop: 8, color: "#bfbfbf", fontSize: 12 }}>
                                    注：更多字段请切换到 JSON 视图查看完整解析结果
                                </div>
                            </>
                        )
                    },
                    {
                        key: "json",
                        label: "JSON视图",
                        children: (
                            <div className="result-panel-scroll">
                                <pre className="result-json">
                                    {JSON.stringify(result.data, null, 2)}
                                </pre>
                            </div>
                        )
                    }
                ]}
            />
        </Card>
    );
}
