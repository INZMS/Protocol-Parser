import { useState } from "react";
import { Card, Select, Input, Button, Space, Tag, message } from "antd";
import { PlayCircleOutlined, ClearOutlined, DownOutlined, UpOutlined } from "@ant-design/icons";

import { useParserStore } from "../../store/parser";
import { mockExamples } from "../../mock/parser";
import { canAccess, useAuthStore } from "../../store/auth";

const MAX_HEX_LENGTH = 4096;
const VISIBLE_EXAMPLES = 3;

export default function InputPanel() {
    const currentUser = useAuthStore((state) => state.user);
    const { protocol, hex, setProtocol, setHex, parse, clear, loading } = useParserStore();
    const [showAllExamples, setShowAllExamples] = useState(false);

    const byteCount = hex ? Math.ceil(hex.replace(/\s/g, "").length / 2) : 0;
    const protocolExamples = mockExamples.filter((example) => example.protocol === protocol);
    const examples = showAllExamples ? protocolExamples : protocolExamples.slice(0, VISIBLE_EXAMPLES);

    const handleProtocolChange = (nextProtocol: string) => {
        setProtocol(nextProtocol);
        setHex("");
        setShowAllExamples(false);
    };

    return (
        <Card title="输入报文" size="small">
            {/* 协议选择 */}
            <div style={{ marginBottom: 14 }}>
                <div style={{ marginBottom: 6, fontWeight: 500, fontSize: 13 }}>选择协议</div>

                <Select
                    style={{ width: "100%" }}
                    value={protocol || undefined}
                    placeholder="请选择协议"
                    onChange={handleProtocolChange}
                    options={[
                        { label: "2929协议", value: "2929" },
                        { label: "JT-808协议", value: "JT-808" },
                        { label: "VDF私有协议", value: "VDF" }
                    ]}
                />
            </div>

            {/* HEX输入 */}
            <div>
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 6 }}>
                    <span style={{ fontWeight: 500, fontSize: 13 }}>报文输入（HEX）</span>

                    <Button type="text" size="small" icon={<ClearOutlined />} onClick={clear}>
                        清空
                    </Button>
                </div>

                <Input.TextArea
                    className="hex-textarea"
                    rows={6}
                    style={{ height: 142, resize: "none" }}
                    maxLength={MAX_HEX_LENGTH}
                    value={hex}
                    onChange={(e) => setHex(e.target.value)}
                    placeholder={!protocol
                        ? "请先选择协议"
                        : protocol === "JT-808"
                            ? "请输入JT-808 HEX报文，例如：\n\n7E02000022013800138000...7E"
                            : protocol === "VDF"
                                ? "请输入VDF ASCII报文对应的HEX，例如：\n\n2A48513230...23"
                                : "请输入2929 HEX报文，例如：\n\n292980013B14941494...0D"}
                />

                <div
                    style={{
                        display: "flex",
                        justifyContent: "space-between",
                        marginTop: 6,
                        color: "#999",
                        fontSize: 12
                    }}
                >
                    <span>共 {byteCount} Bytes</span>
                    <span>{hex.length} / {MAX_HEX_LENGTH}</span>
                </div>
            </div>

            {/* 示例 */}
            {protocol && <div style={{ marginTop: 16 }}>
                <div style={{ fontWeight: 500, marginBottom: 8, fontSize: 13 }}>报文示例</div>

                <Space size={[8, 8]} wrap>
                    {examples.map((example) => (
                        <Tag
                            key={example.label}
                            style={{ cursor: "pointer", marginInlineEnd: 0 }}
                            onClick={() => {
                                setHex(example.hex);
                            }}
                        >
                            {example.label}
                        </Tag>
                    ))}
                </Space>

                {protocolExamples.length > VISIBLE_EXAMPLES && (
                    <div style={{ marginTop: 6 }}>
                        <Button
                            type="link"
                            size="small"
                            style={{ padding: 0, fontSize: 12 }}
                            icon={showAllExamples ? <UpOutlined /> : <DownOutlined />}
                            iconPosition="end"
                            onClick={() => setShowAllExamples((v) => !v)}
                        >
                            {showAllExamples ? "收起" : "更多"}
                        </Button>
                    </div>
                )}
            </div>}

            {/* 解析按钮 */}
            {canAccess(currentUser, "workbench:parser:analyze") && <div style={{ marginTop: "auto", paddingTop: 10 }}>
                <Button
                    type="primary"
                    block
                    icon={<PlayCircleOutlined />}
                    disabled={!protocol || !hex.trim()}
                    loading={loading}
                    onClick={() => parse().catch(() => message.error(useParserStore.getState().error || "解析失败"))}
                >
                    解析报文
                </Button>
            </div>}
        </Card>
    );
}
