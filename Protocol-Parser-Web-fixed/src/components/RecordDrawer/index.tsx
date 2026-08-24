import { Drawer, Descriptions, Divider, Table, Button, message } from "antd";
import { CopyOutlined } from "@ant-design/icons";

import type { HistoryDetail } from "../../pages/Parser/HistoryTable";
import type { ParseField } from "../../store/parser";

interface Props {
    open: boolean;
    data: HistoryDetail | null;
    onClose: () => void;
}
export default function RecordDrawer({ open, data, onClose }: Props) {
    const columns = [
        { title: "字段名称", dataIndex: "name", key: "name", width: 118 },
        { title: "原始值(HEX)", dataIndex: "raw", key: "raw", width: "25%", render: (value: string) => <span className="drawer-value mono">{value}</span> },
        { title: "解析值", dataIndex: "value", key: "value", width: "27%", render: (value: string) => <span className="drawer-value drawer-parsed-value">{value}</span> },
        { title: "说明", dataIndex: "description", key: "description", render: (value: string) => <span className="drawer-value drawer-description">{value}</span> },
        {
            title: "操作", key: "action", width: 54,
            render: (_: unknown, record: ParseField) => (
                <Button type="text" size="small" icon={<CopyOutlined />} aria-label={`复制${record.name}`} onClick={async () => {
                    await navigator.clipboard.writeText(String(record.value ?? ""));
                    message.success(`${record.name} 已复制`);
                }} />
            )
        }
    ];

    return (
        <Drawer
            title="解析记录详情"
            placement="right"
            width="min(920px, 92vw)"
            className="record-detail-drawer"
            open={open}
            onClose={onClose}
        >
            {data && <>
                <Descriptions column={2} size="small">
                    <Descriptions.Item label="协议">{data.protocol}</Descriptions.Item>
                    <Descriptions.Item label="消息ID">{data.messageId}</Descriptions.Item>
                    <Descriptions.Item label="消息名称">{data.messageName}</Descriptions.Item>
                    <Descriptions.Item label="解析时间">{data.time}</Descriptions.Item>
                    <Descriptions.Item label="报文长度">{data.length} Bytes</Descriptions.Item>
                </Descriptions>
                <Divider />
                <section className="drawer-section-heading">
                    <h4>原始报文（HEX）</h4>
                    <Button size="small" icon={<CopyOutlined />} onClick={async () => {
                        await navigator.clipboard.writeText(data.raw);
                        message.success("HEX已复制");
                    }}>复制HEX</Button>
                </section>
                <pre className="drawer-code">{data.raw}</pre>
                <h4>字段解析结果</h4>
                <Table
                    size="small" pagination={false} rowKey={(record) => `${record.index}-${record.offset}`}
                    tableLayout="fixed" columns={columns} dataSource={data.fields}
                />
                <Divider />
                <section className="drawer-section-heading">
                    <h4>业务数据（JSON）</h4>
                    <Button size="small" icon={<CopyOutlined />} onClick={async () => {
                        await navigator.clipboard.writeText(JSON.stringify(data.data, null, 2));
                        message.success("JSON已复制");
                    }}>复制JSON</Button>
                </section>
                <pre className="drawer-code drawer-json">{JSON.stringify(data.data, null, 2)}</pre>
            </>}
        </Drawer>
    );
}
