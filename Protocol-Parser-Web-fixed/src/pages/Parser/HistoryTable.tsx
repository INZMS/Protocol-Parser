import { useCallback, useEffect, useState } from "react";
import axios from "axios";
import { Card, Table, Button, Tag, Input, Space, message, Popconfirm } from "antd";
import { SearchOutlined, ReloadOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";

import RecordDrawer from "../../components/RecordDrawer";
import { useParserStore, type ParseResult } from "../../store/parser";
import { canAccess, useAuthStore } from "../../store/auth";
import { withDateColumnSorters } from "../../utils/tableSorters";

interface HistoryItem {
    id: number;
    time: string;
    protocol: string;
    messageId: string;
    messageName: string;
    length: number;
}
export interface HistoryDetail extends HistoryItem, ParseResult {}

interface HistoryPage {
    items: HistoryItem[];
    total: number;
    page: number;
    pageSize: number;
}

const PAGE_SIZE = 5;

function errorText(error: unknown) {
    return axios.isAxiosError(error)
        ? String(error.response?.data?.error ?? error.message)
        : "历史记录操作失败";
}

export default function HistoryTable() {
    const currentUser = useAuthStore((state) => state.user);
    const canViewHistory = canAccess(currentUser, "workbench:parser:history");
    const historyVersion = useParserStore((state) => state.historyVersion);
    const bumpHistoryVersion = useParserStore((state) => state.bumpHistoryVersion);
    const [drawerOpen, setDrawerOpen] = useState(false);
    const [current, setCurrent] = useState<HistoryDetail | null>(null);
    const [keyword, setKeyword] = useState("");
    const [queryKeyword, setQueryKeyword] = useState("");
    const [page, setPage] = useState(1);
    const [items, setItems] = useState<HistoryItem[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const timer = window.setTimeout(() => {
            setPage(1);
            setQueryKeyword(keyword.trim());
        }, 300);
        return () => window.clearTimeout(timer);
    }, [keyword]);

    const loadHistory = useCallback(async () => {
        if (!canViewHistory) return;
        setLoading(true);
        try {
            const response = await axios.get<{ data: HistoryPage }>("/api/parser/history", {
                params: { page, pageSize: PAGE_SIZE, keyword: queryKeyword }
            });
            setItems(response.data.data.items);
            setTotal(response.data.data.total);
        } catch (error) {
            message.error(errorText(error));
        } finally {
            setLoading(false);
        }
    }, [canViewHistory, page, queryKeyword]);

    useEffect(() => {
        if (!canViewHistory) return;
        void loadHistory();
    }, [canViewHistory, loadHistory, historyVersion]);

    const getDetail = async (id: number) => {
        const response = await axios.get<{ data: HistoryDetail }>(`/api/parser/history/${id}`);
        return response.data.data;
    };

    const columns: ColumnsType<HistoryItem> = [
        { title: "解析时间", dataIndex: "time", key: "time" },
        {
            title: "协议", dataIndex: "protocol", key: "protocol",
            render: (value: string) => <Tag color="blue">{value}</Tag>
        },
        { title: "消息ID", dataIndex: "messageId", key: "messageId" },
        { title: "消息名称", dataIndex: "messageName", key: "messageName" },
        {
            title: "报文长度", dataIndex: "length", key: "length",
            render: (value: number) => `${value} Bytes`
        },
        {
            title: "操作", key: "action", width: 160,
            render: (_, record) => (
                <Space size={4}>
                    {canAccess(currentUser, "workbench:parser:history") && <Button type="link" size="small" style={{ padding: 0 }} onClick={async () => {
                        try {
                            setCurrent(await getDetail(record.id));
                            setDrawerOpen(true);
                        } catch (error) { message.error(errorText(error)); }
                    }}>查看</Button>}
                    {canAccess(currentUser, "workbench:parser:copy") && <Button type="link" size="small" style={{ padding: 0 }} onClick={async () => {
                        try {
                            const detail = await getDetail(record.id);
                            await navigator.clipboard.writeText(JSON.stringify(detail, null, 2));
                            message.success("完整记录已复制");
                        } catch (error) { message.error(errorText(error)); }
                    }}>复制</Button>}
                    {canAccess(currentUser, "workbench:parser:history-delete") && <Popconfirm
                        title="删除这条解析记录？" okText="删除" cancelText="取消"
                        onConfirm={async () => {
                            try {
                                await axios.delete(`/api/parser/history/${record.id}`);
                                message.success("记录已删除");
                                if (items.length === 1 && page > 1) setPage(page - 1);
                                else bumpHistoryVersion();
                            } catch (error) { message.error(errorText(error)); }
                        }}
                    >
                        <Button type="link" size="small" danger style={{ padding: 0 }}>删除</Button>
                    </Popconfirm>}
                </Space>
            )
        }
    ];

    if (!canViewHistory) return null;

    return (
        <>
            <Card className="history-card" title="解析记录（历史）" size="small" extra={
                <Space>
                    <Input
                        allowClear size="small" placeholder="搜索消息ID/名称/协议"
                        prefix={<SearchOutlined />} value={keyword}
                        onChange={(event) => setKeyword(event.target.value)} style={{ width: 220 }}
                    />
                    <Button size="small" type="primary" icon={<ReloadOutlined />} onClick={() => void loadHistory()}>
                        刷新
                    </Button>
                </Space>
            }>
                <Table
                    className="history-table"
                    columns={withDateColumnSorters<HistoryItem>(columns as Array<Record<string, any>>)} dataSource={items} rowKey="id" size="small" loading={loading}
                    pagination={{
                        current: page, pageSize: PAGE_SIZE, total, size: "small",
                        showTotal: (value) => `共 ${value} 条`, onChange: setPage
                    }}
                />
            </Card>
            <RecordDrawer open={drawerOpen} data={current} onClose={() => setDrawerOpen(false)} />
        </>
    );
}
