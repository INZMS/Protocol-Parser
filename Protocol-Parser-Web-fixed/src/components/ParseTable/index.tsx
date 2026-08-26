import { Button, Table, Tooltip, Typography, message } from "antd";
import { CopyOutlined } from "@ant-design/icons";

import { useParserStore } from "../../store/parser";
import { canAccess, useAuthStore } from "../../store/auth";

export default function ParseTable() {
    const currentUser = useAuthStore(state => state.user);
    const { result } = useParserStore();

    if (!result) {
        return null;
    }

    const data = result.fields.map((item, index) => ({ key: index, ...item }));

    const columns = [
        {
            title: "字段信息",
            dataIndex: "name",
            key: "name",
            width: "34%",
            render: (value: string, record: any) => (
                <div className="parse-field-info">
                    <div className="parse-field-heading">
                        <span className="parse-field-name">{value}</span>
                        <span className="parse-position">
                            偏移 {record.offset} · {record.length}B
                        </span>
                    </div>
                    <span className="parse-description">{record.description}</span>
                </div>
            )
        },
        {
            title: "原始值（HEX）",
            dataIndex: "raw",
            key: "raw",
            width: "27%",
            render: (value: string) => (
                <Typography.Text className="mono parse-cell-raw">
                    {String(value ?? "")}
                </Typography.Text>
            )
        },
        {
            title: "解析结果",
            dataIndex: "value",
            key: "value",
            width: "39%",
            render: (value: string) => (
                <div className="parse-value-with-action">
                    <Typography.Text className="parse-cell-value">
                        {String(value ?? "")}
                    </Typography.Text>
                    <Tooltip title="复制完整解析值">
                        {canAccess(currentUser, "workbench:parser:copy") && <Button
                            className="parse-copy-button"
                            type="text"
                            size="small"
                            icon={<CopyOutlined />}
                            aria-label="复制完整解析值"
                            onClick={() => {
                                navigator.clipboard.writeText(String(value ?? ""));
                                message.success("解析值已复制");
                            }}
                        />}
                    </Tooltip>
                </div>
            )
        }
    ];

    return (
        <Table
            className="parse-result-table"
            columns={columns}
            dataSource={data}
            pagination={false}
            size="small"
            tableLayout="fixed"
            rowClassName={(_, index) => index % 2 === 1 ? "parse-row-even" : ""}
        />
    );
}
