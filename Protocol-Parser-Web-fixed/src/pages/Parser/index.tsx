import { useEffect, useMemo, useState, type ReactNode } from "react";
import axios from "axios";
import { Breadcrumb, Layout, Menu, type MenuProps } from "antd";
import { ApartmentOutlined, BankOutlined, CarOutlined, CodeOutlined, DatabaseOutlined, DesktopOutlined, FundOutlined, HddOutlined, InboxOutlined, MenuOutlined, SettingOutlined, TeamOutlined, ToolOutlined, UserOutlined } from "@ant-design/icons";
import Header from "../../components/AppHeader";
import InputPanel from "./InputPanel";
import ResultPanel from "./ResultPanel";
import HistoryTable from "./HistoryTable";
import UserManagement from "../Admin/UserManagement";
import RoleManagement from "../Admin/RoleManagement";
import MenuManagement from "../Admin/MenuManagement";
import SystemSettings from "../SystemSettings";
import VehicleManagement from "../BasicInfo/VehicleManagement";
import DeviceInventory from "../BasicInfo/DeviceInventory";
import DeviceMaintenance from "../BasicInfo/DeviceMaintenance";
import OrganizationManagement from "../BasicInfo/OrganizationManagement";
import FinanceCompanyManagement from "../BasicInfo/FinanceCompanyManagement";
import FinanceProductManagement from "../BasicInfo/FinanceProductManagement";
import CollectionCompanyManagement from "../BasicInfo/CollectionCompanyManagement";
import ExceptionPage from "../../components/ExceptionPage";
import { useAuthStore } from "../../store/auth";
import { useSettingsStore } from "../../store/settings";

const { Content, Sider } = Layout;
type PageKey = "parser" | "organizations" | "vehicles" | "deviceInventory" | "deviceMaintenance" | "financeCompanies" | "financeProducts" | "collectionCompanies" | "users" | "roles" | "menus" | "settings";
const breadcrumbMap: Record<PageKey, string[]> = {
    parser: ["工作台", "协议解析"], organizations:["基础信息","组织机构管理"], vehicles: ["基础信息", "车辆综合管理"], deviceInventory: ["基础信息", "设备管理", "设备库存管理"], deviceMaintenance: ["基础信息", "设备管理", "设备生命周期管理"], users: ["系统管理", "用户管理"], roles: ["系统管理", "角色管理"],
    financeCompanies:["基础信息","合作方管理","金融公司管理"],financeProducts:["基础信息","合作方管理","金融产品管理"],collectionCompanies:["基础信息","合作方管理","清收公司管理"], menus: ["系统管理", "菜单管理"], settings: ["系统管理", "系统设置"]
};
const menuCodeByKey:Record<string,string>={workbench:"workbench",parser:"workbench:parser",basic:"basic",organizations:"basic:organization",vehicles:"basic:vehicle",device:"basic:device",deviceInventory:"basic:device:inventory",deviceMaintenance:"basic:device:maintenance",partner:"basic:partner",financeCompanies:"basic:partner:finance-company",financeProducts:"basic:partner:finance-product",collectionCompanies:"basic:partner:collection-company",system:"system",users:"system:user",roles:"system:role",menus:"system:menu",settings:"system:settings"};
const pageOrder:PageKey[]=["parser","organizations","vehicles","deviceInventory","deviceMaintenance","financeCompanies","financeProducts","collectionCompanies","users","roles","menus","settings"];
const ACTIVE_PAGE_KEY="protocol_parser_active_page";
type MenuConfig={code:string;name:string;status:number;sortOrder:number};
const applyMenuConfig=(items:MenuProps["items"],config:Record<string,MenuConfig>,permissions:Set<string>):MenuProps["items"]=>
    (items||[]).flatMap(item=>{if(!item)return[];const value=item as any;const code=menuCodeByKey[String(value.key)];const current=code?config[code]:undefined;const permitted=!code||permissions.has(code)||[...permissions].some(itemCode=>itemCode.startsWith(`${code}:`));if(current?.status===0||!permitted)return[];const children=value.children?applyMenuConfig(value.children,config,permissions):undefined;return[{...value,label:current?.name||value.label,children,_sortOrder:current?.sortOrder??0}]}).sort((a:any,b:any)=>(a?._sortOrder??0)-(b?._sortOrder??0));

export default function Parser() {
    const user = useAuthStore((state) => state.user);
    const navigationType = useSettingsStore((state) => state.loginPage.navigationType);
    const [activePage, setActivePage] = useState<PageKey>(()=>{
        const saved=localStorage.getItem(ACTIVE_PAGE_KEY);
        return pageOrder.includes(saved as PageKey)?saved as PageKey:"parser";
    });
    const [menuConfig,setMenuConfig]=useState<Record<string,MenuConfig>>({});
    const [mobileNav,setMobileNav]=useState(false);
    const [collapsed, setCollapsed] = useState(() => localStorage.getItem("protocol_parser_sidebar_collapsed") === "1");
    const baseMenuItems = useMemo<MenuProps["items"]>(() => [
        { key: "workbench", icon: <DesktopOutlined />, label: "工作台", children: [{ key: "parser", icon: <CodeOutlined />, label: "协议解析" }] },
        { key: "basic", icon: <DatabaseOutlined />, label: "基础信息", children: [
            { key: "organizations", icon: <ApartmentOutlined />, label: "组织机构管理" },
            { key: "vehicles", icon: <CarOutlined />, label: "车辆综合管理" },
            { key: "device", icon: <HddOutlined />, label: "设备管理", children: [
                { key: "deviceInventory", icon: <InboxOutlined />, label: "设备库存管理" },
                { key: "deviceMaintenance", icon: <ToolOutlined />, label: "设备生命周期管理" }
            ] }
            ,{key:"partner",icon:<TeamOutlined/>,label:"合作方管理",children:[{key:"financeCompanies",icon:<BankOutlined/>,label:"金融公司管理"},{key:"financeProducts",icon:<FundOutlined/>,label:"金融产品管理"},{key:"collectionCompanies",icon:<TeamOutlined/>,label:"清收公司管理"}]}
        ] },
        ...(((user?.permissions||[]).some(code=>code==="system"||code.startsWith("system:"))) ? [{ key: "system", icon: <SettingOutlined />, label: "系统管理", children: [
            { key: "users", icon: <UserOutlined />, label: "用户管理" }, { key: "roles", icon: <TeamOutlined />, label: "角色管理" },
            { key: "menus", icon: <MenuOutlined />, label: "菜单管理" }, { key: "settings", icon: <SettingOutlined />, label: "系统设置" }
        ] }] : [])
    ], [user?.permissions]);
    const permissionSet=useMemo(()=>new Set(user?.permissions||[]),[user?.permissions]);
    const canOpenPage=(key:PageKey)=>{
        const code=menuCodeByKey[key];
        return menuConfig[code]?.status!==0&&(permissionSet.has(code)||[...permissionSet].some(permission=>permission.startsWith(`${code}:`)));
    };
    const menuItems=useMemo(()=>applyMenuConfig(baseMenuItems,menuConfig,permissionSet),[baseMenuItems,menuConfig,permissionSet]);
    useEffect(()=>{if(!(user?.permissions||[]).includes("system:menu"))return;const loadMenuConfig=async()=>{try{const response=await axios.get("/api/admin/menus",{params:{page:1,pageSize:500}});setMenuConfig(Object.fromEntries((response.data.menus||[]).map((item:MenuConfig)=>[item.code,item])))}catch{/* 菜单配置加载失败时保留内置导航 */}};void loadMenuConfig();window.addEventListener("menu-config-changed",loadMenuConfig);return()=>window.removeEventListener("menu-config-changed",loadMenuConfig)},[user?.permissions]);
    useEffect(()=>{
        if(!user)return;
        if(canOpenPage(activePage))return;
        const fallback=pageOrder.find(canOpenPage);
        if(fallback){setActivePage(fallback);localStorage.setItem(ACTIVE_PAGE_KEY,fallback)}
    },[activePage,menuConfig,user?.id,user?.permissions]);
    const navigate: MenuProps["onClick"] = ({ key }) => {
        const page=key as PageKey;
        if(!pageOrder.includes(page)||!canOpenPage(page))return;
        setActivePage(page);
        localStorage.setItem(ACTIVE_PAGE_KEY,page);
    };
    const parserPage = <div className="parser-page">
        <div className="main-grid"><div className="left-panel"><InputPanel /></div><div className="right-panel"><ResultPanel /></div></div>
        <HistoryTable />
    </div>;
    const pages: Record<PageKey, ReactNode> = { parser: parserPage, organizations:<OrganizationManagement/>, vehicles: <VehicleManagement />, deviceInventory: <DeviceInventory />, deviceMaintenance: <DeviceMaintenance />, financeCompanies:<FinanceCompanyManagement/>,financeProducts:<FinanceProductManagement/>,collectionCompanies:<CollectionCompanyManagement/>, users: <UserManagement />, roles: <RoleManagement />, menus: <MenuManagement />, settings: <SystemSettings /> };
    const toggleSidebar=()=>{const value=!collapsed;setCollapsed(value);localStorage.setItem("protocol_parser_sidebar_collapsed",value?"1":"0")};
    return <Layout className="admin-shell" hasSider={navigationType==="sidebar"}>
            {navigationType === "sidebar" && <Sider className="admin-sider dawn-blue-sider" theme="dark" breakpoint="xl" onBreakpoint={broken=>{setMobileNav(broken);if(broken)setCollapsed(true)}} collapsed={collapsed} trigger={null} width={224} collapsedWidth={mobileNav?0:64}>
                <div className="pro-sidebar-brand"><div className="pro-sidebar-logo"><img src="/favicon.png" alt="系统标志"/></div>{!collapsed&&<strong>协议解析工具</strong>}</div>
                <Menu theme="dark" mode="inline" items={menuItems} selectedKeys={[activePage]} defaultOpenKeys={["workbench", "basic", "device", "partner", "system"]} onClick={navigate} />
            </Sider>}
            <Layout className="admin-main">
                <Header showBrand={navigationType==="top"} sidebarCollapsed={collapsed} onSidebarToggle={toggleSidebar}/>
                {navigationType === "top" && <div className="admin-top-menu"><Menu mode="horizontal" items={menuItems} selectedKeys={[activePage]} onClick={navigate} /></div>}
                <div className="admin-breadcrumb"><Breadcrumb items={breadcrumbMap[activePage].map(title => ({ title }))} /></div>
                <Content className={`admin-content page-${activePage}`}>{!(activePage in pages)?<ExceptionPage code={404}/>:(!user||!canOpenPage(activePage))?<ExceptionPage code={403}/>:pages[activePage]}</Content>
                <footer className="admin-global-footer">© 2026 协议解析工具 · 让协议解析更简单高效</footer>
            </Layout>
    </Layout>;
}
