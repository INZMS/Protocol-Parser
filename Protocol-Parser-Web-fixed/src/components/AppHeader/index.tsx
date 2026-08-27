import { useEffect, useState } from "react";
import axios from "axios";
import {
    DownOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    QuestionCircleOutlined,
    SafetyCertificateOutlined,
    BellOutlined,
    UserOutlined
} from "@ant-design/icons";
import { Avatar, Badge, Button, Descriptions, Drawer, Dropdown, Form, Input, Modal, Tag, message, type MenuProps } from "antd";

import { useParserStore } from "../../store/parser";
import { useAuthStore } from "../../store/auth";
import { useSettingsStore } from "../../store/settings";
export default function Header({showBrand=true,sidebarCollapsed=false,onSidebarToggle}:{showBrand?:boolean;sidebarCollapsed?:boolean;onSidebarToggle?:()=>void}) {
    type NotificationItem={id:number;title:string;content:string;organizationName:string;isRead:boolean;createdAt:string};
    const [helpOpen, setHelpOpen] = useState(false);
    const [profileOpen, setProfileOpen] = useState(false);
    const [logoutOpen, setLogoutOpen] = useState(false);
    const [logoutLoading, setLogoutLoading] = useState(false);
    const [notifications,setNotifications]=useState<NotificationItem[]>([]);
    const { clear, setProtocol } = useParserStore();
    const { user, updateProfile, logout } = useAuthStore();
    const branding = useSettingsStore((state) => state.loginPage);
    const [profileForm] = Form.useForm();

    const loadNotifications=async()=>{try{const response=await axios.get("/api/basic-info/notifications");setNotifications(response.data.notifications||[])}catch{/* 登录初期或无权限时静默忽略 */}};
    useEffect(()=>{void loadNotifications();const timer=window.setInterval(()=>void loadNotifications(),60000);return()=>window.clearInterval(timer)},[]);
    const notificationItems:MenuProps["items"]=(notifications.length?notifications.slice(0,10).map(item=>({key:String(item.id),label:<div style={{width:360,whiteSpace:"normal"}}><div style={{fontWeight:item.isRead?400:600}}>{item.title}{item.organizationName?` · ${item.organizationName}`:""}</div><div style={{color:"#64748b",marginTop:4}}>{item.content}</div><div style={{fontSize:12,color:"#94a3b8",marginTop:4}}>{item.createdAt}</div></div>})):[{key:"empty",disabled:true,label:"暂无站内通知"}]);

    const userMenu: MenuProps["items"] = [
        {
            key: "profile",
            icon: <UserOutlined />,
            label: "个人信息"
        },
        { type: "divider" },
        {
            key: "logout",
            icon: <LogoutOutlined />,
            label: "退出登录",
            danger: true
        }
    ];

    const handleUserMenu: MenuProps["onClick"] = ({ key }) => {
        if (key === "profile") {
            profileForm.setFieldsValue({ displayName: user?.displayName, email: user?.email, phone: user?.phone });
            setProfileOpen(true);
        } else if (key === "logout") {
            setLogoutOpen(true);
        }
    };


    const handleLogout = async () => {
        setLogoutLoading(true);
        try {
            await logout();
            clear();
            setProtocol("");
            setLogoutOpen(false);
            message.success("退出成功，期待再次使用");
        } finally {
            setLogoutLoading(false);
        }
    };

    const saveProfile = async () => {
        const values = await profileForm.validateFields();
        await updateProfile(values);
        message.success("个人资料已更新");
        setProfileOpen(false);
    };

    return (
        <>
        <div className="pro-global-header"
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
            {/* 左侧品牌或菜单折叠控制 */}
            <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
                {!showBrand&&<button className="pro-sidebar-toggle" type="button" onClick={onSidebarToggle} aria-label={sidebarCollapsed?"展开菜单":"收起菜单"}>{sidebarCollapsed?<MenuUnfoldOutlined/>:<MenuFoldOutlined/>}</button>}
                {showBrand&&<>
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
                    <img src={branding.systemIcon} alt="系统图标" style={{width:24,height:24,objectFit:"contain"}} />
                </div>

                <div style={{ minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, lineHeight: "20px", whiteSpace: "nowrap" }}>
                        {branding.systemName}
                    </div>
                    <div style={{ fontSize: 11, color: "#8c8c8c", lineHeight: "16px" }}>
                        {branding.systemNameEn}
                    </div>
                </div>
                </>}
            </div>

            {/* 右侧 */}
            <div style={{ display: "flex", alignItems: "center", gap: 20, flex: "0 0 auto", fontSize: 13 }}>
                <Dropdown placement="bottomRight" trigger={["click"]} menu={{items:notificationItems,onClick:async({key})=>{if(key!=="empty"){await axios.put(`/api/basic-info/notifications/${key}/read`);void loadNotifications()}}}}>
                    <button className="header-help-button" type="button" aria-label="站内通知"><Badge count={notifications.filter(item=>!item.isRead).length} size="small"><BellOutlined style={{fontSize:16}}/></Badge><span>通知</span></button>
                </Dropdown>
                <button className="header-help-button" type="button" onClick={() => setHelpOpen(true)}>
                    <QuestionCircleOutlined />
                    <span>使用说明</span>
                </button>

                <Dropdown menu={{ items: userMenu, onClick: handleUserMenu }} placement="bottomRight" trigger={["click"]}>
                    <button className="header-user-button" type="button" aria-label="打开用户菜单">
                        <Avatar size={30} style={{ background: "#e6f4ff", color: "#1677ff" }}>{(user?.displayName || user?.username || "A").slice(0, 1).toUpperCase()}</Avatar>
                        <span className="header-user-name">{user?.displayName || user?.username}</span>
                        <DownOutlined className="header-user-arrow" />
                    </button>
                </Dropdown>
            </div>
        </div>

        <Drawer
            title={`${branding.systemName} · 使用说明`}
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
                        当前支持 2929、JT-808 和 VDF 私有协议。2929重点解析通用头和位置上报（0x80），包括定位时间（北京时间）、经纬度换算、
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

        <Drawer
            title="个人信息"
            placement="right"
            width="min(520px, 92vw)"
            open={profileOpen}
            onClose={() => setProfileOpen(false)}
            extra={<Button type="primary" onClick={() => void saveProfile()}>保存</Button>}
        >
            <div className="profile-page">
                <div className="profile-summary">
                    <Avatar size={72} style={{ background: "#1677ff", fontSize: 28 }}>{(user?.displayName || user?.username || "A").slice(0, 1).toUpperCase()}</Avatar>
                    <div>
                        <h2>{user?.displayName || user?.username}</h2>
                        <div className="profile-role">
                            <SafetyCertificateOutlined />
                            系统管理员
                        </div>
                    </div>
                </div>

                <Descriptions className="profile-details" column={1} bordered size="middle">
                    <Descriptions.Item label="用户名">{user?.username}</Descriptions.Item>
                    <Descriptions.Item label="角色"><Tag color="blue">{user?.role === "admin" ? "管理员" : user?.role}</Tag></Descriptions.Item>
                    <Descriptions.Item label="所属系统">{branding.systemName}</Descriptions.Item>
                    <Descriptions.Item label="账号状态"><Tag color="success">正常</Tag></Descriptions.Item>
                </Descriptions>

                <Form form={profileForm} layout="vertical" className="profile-form">
                    <Form.Item label="显示名称" name="displayName" rules={[{ required: true, message: "请输入显示名称" }]}><Input maxLength={64} /></Form.Item>
                    <Form.Item label="邮箱" name="email" rules={[{ type: "email", message: "邮箱格式不正确" }]}><Input placeholder="可选" maxLength={128} /></Form.Item>
                    <Form.Item label="手机号" name="phone"><Input placeholder="可选" maxLength={32} /></Form.Item>
                </Form>
            </div>
        </Drawer>

        <Modal
            title="退出登录"
            open={logoutOpen}
            closable={!logoutLoading}
            maskClosable={!logoutLoading}
            onCancel={() => {if(!logoutLoading)setLogoutOpen(false)}}
            footer={[
                <Button key="cancel" disabled={logoutLoading} onClick={() => setLogoutOpen(false)}>取消</Button>,
                <Button key="logout" type="primary" danger loading={logoutLoading} icon={<LogoutOutlined />} onClick={() => void handleLogout()}>确认退出</Button>
            ]}
        >
            <p>确定退出当前账号吗？系统将安全结束本次登录，已保存的业务数据和解析历史不会受到影响。</p>
        </Modal>
        </>
    );
}
