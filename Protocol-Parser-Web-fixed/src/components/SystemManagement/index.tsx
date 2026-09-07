import { useEffect, useMemo, useState, type Key } from "react";
import axios from "axios";
import { Alert, Button, Checkbox, Drawer, Form, Input, InputNumber, Popconfirm, Radio, Select, Space, Tag, Tree, message } from "antd";
import { DeleteOutlined, EditOutlined, KeyOutlined, PlusOutlined, ReloadOutlined, SafetyCertificateOutlined, SearchOutlined } from "@ant-design/icons";
import { canAccess, useAuthStore } from "../../store/auth";
import { withDateColumnSorters } from "../../utils/tableSorters";
import { getErrorMessage } from "../../api/client";
import DataGrid from "../DataGrid";
import ListPageShell from "../ListPageShell";
import EnableStatusSwitch from "../EnableStatusSwitch";
type Role = {
    id: number;
    code: string;
    name: string;
    description: string;
    status: number;
    isSystem: boolean;
    menuIds: number[];
    organizationId?: number;
    organizationName: string;
    createdBy: string;
    createdAt: string;
};
type User = {
    id: number;
    username: string;
    displayName: string;
    email: string;
    phone: string;
    status: number;
    roleId?: number;
    roleName: string;
    organizationId?: number;
    organizationName: string;
    createdBy: string;
    createdAt: string;
};
type Organization = {
    id: number;
    name: string;
    code: string;
    status: number;
};
type Menu = {
    id: number;
    parentId: number;
    name: string;
    code: string;
    menuType: "directory" | "menu" | "button";
    path: string;
    icon: string;
    sortOrder: number;
    status: number;
    createdBy: string;
    createdAt: string;
    children?: Menu[];
};
type Page = "users" | "roles" | "menus";
const errorText = (e: unknown) => getErrorMessage(e);
const toTree = (items: Menu[], parentId = 0): Menu[] => items.filter(i => i.parentId === parentId).map(i => ({ ...i, children: toTree(items, i.id) }));
const typeMeta = { directory: { text: "目录", color: "purple" }, menu: { text: "菜单", color: "blue" }, button: { text: "按钮", color: "orange" } } as const;
export default function SystemManagement({ page }: {
    page: Page;
}) {
    const currentUser = useAuthStore(state => state.user), refreshPermissions = useAuthStore(state => state.refresh), can = (code: string) => canAccess(currentUser, code);
    const [users, setUsers] = useState<User[]>([]), [roles, setRoles] = useState<Role[]>([]), [menus, setMenus] = useState<Menu[]>([]), [organizations, setOrganizations] = useState<Organization[]>([]), [loading, setLoading] = useState(false);
    const [listPage, setListPage] = useState(1), [listPageSize, setListPageSize] = useState(10), [listTotal, setListTotal] = useState(0), [searchVersion, setSearchVersion] = useState(0);
    const [userOpen, setUserOpen] = useState(false), [roleCreateOpen, setRoleCreateOpen] = useState(false), [roleEditOpen, setRoleEditOpen] = useState(false), [menuOpen, setMenuOpen] = useState(false);
    const [passwordUser, setPasswordUser] = useState<User | null>(null), [permissionRole, setPermissionRole] = useState<Role | null>(null);
    const [userForm] = Form.useForm(), [roleCreateForm] = Form.useForm(), [roleEditForm] = Form.useForm(), [menuForm] = Form.useForm(), [passwordForm] = Form.useForm();
    const [checkedMenus, setCheckedMenus] = useState<Key[]>([]);
    const [filters, setFilters] = useState({ keyword: "", status: undefined as number | undefined, roleId: undefined as number | undefined, menuType: undefined as Menu["menuType"] | undefined });
    const menuTree = useMemo(() => toTree(menus), [menus]);
    const allMenuIds = useMemo(() => menus.map(m => m.id), [menus]);
    const load = async () => { setLoading(true); try {
        if (page === "users") {
            const requests: Promise<void>[] = [];
            if (can("system:user:query"))
                requests.push(axios.get("/api/admin/users", { params: { page: listPage, pageSize: listPageSize, keyword: filters.keyword || undefined, status: filters.status, roleId: filters.roleId } }).then(response => { setUsers(response.data.users || []); setListTotal(response.data.total || 0); }));
            if (can("system:role:query"))
                requests.push(axios.get("/api/admin/roles", { params: { page: 1, pageSize: 200 } }).then(response => setRoles(response.data.roles || [])));
            await Promise.all(requests);
        }
        else if (page === "roles") {
            const requests: Promise<void>[] = [];
            if (can("system:role:query"))
                requests.push(axios.get("/api/admin/roles", { params: { page: listPage, pageSize: listPageSize, keyword: filters.keyword || undefined, status: filters.status } }).then(response => { setRoles(response.data.roles || []); setListTotal(response.data.total || 0); }));
            if (can("system:menu:query"))
                requests.push(axios.get("/api/admin/menus/tree").then(response => setMenus(response.data.menus || [])));
            await Promise.all(requests);
        }
        else if (can("system:menu:query")) {
            const response = await axios.get("/api/admin/menus/tree", { params: { keyword: filters.keyword || undefined, status: filters.status, menuType: filters.menuType } });
            setMenus(response.data.menus || []);
            setListTotal(response.data.total || 0);
        }
    }
    catch (e) {
        message.error(errorText(e));
    }
    finally {
        setLoading(false);
    } };
    useEffect(() => { void load(); }, [page, listPage, listPageSize, searchVersion]);
    useEffect(() => { if (page === "users" || page === "roles")
        void axios.get("/api/basic-info/organizations", { params: { page: 1, pageSize: 500 } }).then(response => setOrganizations(response.data.organizations || [])).catch(() => setOrganizations([])); }, [page]);
    useEffect(() => { setFilters({ keyword: "", status: undefined, roleId: undefined, menuType: undefined }); setListPage(1); setListPageSize(10); setListTotal(0); }, [page]);
    const saveUser = async () => { try {
        const v = await userForm.validateFields();
        v.id ? await axios.put(`/api/admin/users/${v.id}`, v) : await axios.post("/api/admin/users", v);
        setUserOpen(false);
        message.success("用户已保存");
        void load();
    }
    catch (e) {
        if (axios.isAxiosError(e))
            message.error(errorText(e));
    } };
    const saveNewRole = async () => { try {
        const v = await roleCreateForm.validateFields();
        await axios.post("/api/admin/roles", v);
        setRoleCreateOpen(false);
        message.success("角色已新增");
        void load();
    }
    catch (e) {
        if (axios.isAxiosError(e))
            message.error(errorText(e));
    } };
    const menuWithAncestors = (keys: Key[]) => { const selected = new Set(keys.map(Number)); const byID = new Map(menus.map(menu => [menu.id, menu])); for (const id of [...selected]) {
        let parent = byID.get(id)?.parentId || 0;
        while (parent) {
            selected.add(parent);
            parent = byID.get(parent)?.parentId || 0;
        }
    } return [...selected]; };
    const saveEditedRole = async () => { try {
        const v = await roleEditForm.validateFields();
        await axios.put(`/api/admin/roles/${v.id}`, v);
        if (can("system:role:permission") && v.id && !(permissionRole?.isSystem && permissionRole.code === "admin"))
            await axios.put(`/api/admin/roles/${v.id}/menus`, { menuIds: menuWithAncestors(checkedMenus) });
        setRoleEditOpen(false);
        setPermissionRole(null);
        await refreshPermissions();
        message.success("角色信息与权限已保存");
        void load();
    }
    catch (e) {
        if (axios.isAxiosError(e))
            message.error(errorText(e));
    } };
    const notifyMenuChanged = () => window.dispatchEvent(new Event("menu-config-changed"));
    const saveMenu = async () => { try {
        const v = await menuForm.validateFields();
        v.id ? await axios.put(`/api/admin/menus/${v.id}`, v) : await axios.post("/api/admin/menus", v);
        setMenuOpen(false);
        message.success("菜单权限项已保存");
        await load();
        notifyMenuChanged();
    }
    catch (e) {
        if (axios.isAxiosError(e))
            message.error(errorText(e));
    } };
    const toggleMenu = async (item: Menu, enabled: boolean) => { try {
        await axios.put(`/api/admin/menus/${item.id}`, { ...item, status: enabled ? 1 : 0 });
        setMenus(current => current.map(menu => menu.id === item.id ? { ...menu, status: enabled ? 1 : 0 } : menu));
        await refreshPermissions();
        notifyMenuChanged();
        message.success(enabled ? "菜单已启用" : "菜单已停用");
    }
    catch (e) {
        message.error(errorText(e));
    } };
    const toggleUser = async (item: User, enabled: boolean) => { try {
        await axios.put(`/api/admin/users/${item.id}`, { ...item, status: enabled ? 1 : 0 });
        setUsers(current => current.map(user => user.id === item.id ? { ...user, status: enabled ? 1 : 0 } : user));
        message.success(enabled ? "用户已启用" : "用户已停用");
    }
    catch (e) {
        message.error(errorText(e));
        throw e;
    } };
    const toggleRole = async (item: Role, enabled: boolean) => { try {
        await axios.put(`/api/admin/roles/${item.id}`, { ...item, status: enabled ? 1 : 0 });
        setRoles(current => current.map(role => role.id === item.id ? { ...role, status: enabled ? 1 : 0 } : role));
        message.success(enabled ? "角色已启用" : "角色已停用");
    }
    catch (e) {
        message.error(errorText(e));
        throw e;
    } };
    const remove = async (type: Page, id: number) => { try {
        await axios.delete(`/api/admin/${type}/${id}`);
        message.success("已删除");
        await load();
        if (type === "menus")
            notifyMenuChanged();
    }
    catch (e) {
        message.error(errorText(e));
    } };
    const permissionTree = (nodes: Menu[]): any[] => nodes.map(n => ({ key: n.id, title: <span>{n.name} <Tag color={typeMeta[n.menuType].color}>{typeMeta[n.menuType].text}</Tag></span>, children: permissionTree(n.children || []) }));
    const menuOptions = menus.filter(m => m.menuType !== "button").map(m => ({ value: m.id, label: `${m.name}（${m.code}）` }));
    const normalizedKeyword = filters.keyword.trim().toLowerCase();
    const matchedMenuIds = useMemo(() => new Set(menus.filter(item => (!normalizedKeyword || [item.name, item.code, item.path].some(value => String(value || "").toLowerCase().includes(normalizedKeyword))) && (filters.status === undefined || item.status === filters.status) && (!filters.menuType || item.menuType === filters.menuType)).map(item => item.id)), [menus, normalizedKeyword, filters.status, filters.menuType]);
    const filteredMenuTree = useMemo(() => { const walk = (nodes: Menu[]): Menu[] => nodes.map(node => { const children = walk(node.children || []); return matchedMenuIds.has(node.id) || children.length ? { ...node, children } : null; }).filter(Boolean) as Menu[]; return walk(menuTree); }, [menuTree, matchedMenuIds]);
    const queryList = () => { setListPage(1); setSearchVersion(value => value + 1); };
    const resetFilters = () => { setFilters({ keyword: "", status: undefined, roleId: undefined, menuType: undefined }); setListPage(1); setSearchVersion(value => value + 1); };
    const listPagination = { current: listPage, pageSize: listPageSize, total: listTotal, showSizeChanger: true, showTotal: (total: number) => `共 ${total} 条`, onChange: (current: number, size: number) => { setListPage(current); setListPageSize(size); } };
    const drawerProps = { width: 560, destroyOnHidden: true };
    const pageContent = {
        users: <ListPageShell className="system-list-shell" query={<div className="query-form-row"><label>用户检索<Input allowClear prefix={<SearchOutlined />} placeholder="用户名 / 姓名 / 手机号 / 邮箱" value={filters.keyword} onChange={event => setFilters(value => ({ ...value, keyword: event.target.value }))}/></label><label>所属角色<Select allowClear placeholder="请选择角色" value={filters.roleId} onChange={roleId => setFilters(value => ({ ...value, roleId }))} options={roles.map(role => ({ value: role.id, label: role.name }))}/></label><label>用户状态<Select allowClear placeholder="请选择状态" value={filters.status} onChange={status => setFilters(value => ({ ...value, status }))} options={[{ value: 1, label: "正常" }, { value: 0, label: "停用" }]}/></label><div className="query-actions">{can("system:user:query") && <Button type="primary" icon={<SearchOutlined />} onClick={queryList}>查询</Button>}<Button icon={<ReloadOutlined />} onClick={resetFilters}>重置</Button></div></div>} actions={<div className="system-action-row">{can("system:user:add") && <Button type="primary" icon={<PlusOutlined />} onClick={() => { userForm.resetFields(); userForm.setFieldsValue({ status: 1 }); setUserOpen(true); }}>新增用户</Button>}</div>}><DataGrid<User> title="用户列表" preferenceKey="list-columns-system-users" rowKey="id" loading={loading} dataSource={users} pagination={listPagination} columns={[
                ...withDateColumnSorters<User>([{ title: "用户名", dataIndex: "username" }, { title: "显示名称", dataIndex: "displayName" }, { title: "角色", dataIndex: "roleName", render: (v: string) => <Tag color="blue">{v || "未分配"}</Tag> }, { title: "所属机构", dataIndex: "organizationName", render: (v: string) => <span>{v || "未绑定"}</span> }, { title: "启用状态", dataIndex: "status", width: 110, render: (v: number, item: User) => <EnableStatusSwitch checked={v === 1} disabled={!can("system:user:edit")} onChange={can("system:user:edit") ? enabled => toggleUser(item, enabled) : undefined}/> }, { title: "创建人", dataIndex: "createdBy", width: 120, ellipsis: true, render: (v: string) => v || "系统管理员" }, { title: "创建时间", dataIndex: "createdAt" }]), ...((can("system:user:edit") || can("system:user:password") || can("system:user:delete")) ? [{ title: "操作", width: 300, render: (_: unknown, v: User) => <Space>{can("system:user:edit") && <Button type="link" icon={<EditOutlined />} onClick={() => { userForm.setFieldsValue(v); setUserOpen(true); }}>编辑</Button>}{can("system:user:password") && <Button type="link" icon={<KeyOutlined />} onClick={() => { passwordForm.resetFields(); setPasswordUser(v); }}>重置密码</Button>}{can("system:user:delete") && <Popconfirm title="确认删除该用户？" onConfirm={() => void remove("users", v.id)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm>}</Space> }] : [])
            ]}/></ListPageShell>,
        roles: <ListPageShell className="system-list-shell" query={<div className="query-form-row"><label>角色检索<Input allowClear prefix={<SearchOutlined />} placeholder="角色名称 / 编码 / 说明" value={filters.keyword} onChange={event => setFilters(value => ({ ...value, keyword: event.target.value }))}/></label><label>角色状态<Select allowClear placeholder="请选择状态" value={filters.status} onChange={status => setFilters(value => ({ ...value, status }))} options={[{ value: 1, label: "正常" }, { value: 0, label: "停用" }]}/></label><div className="query-actions">{can("system:role:query") && <Button type="primary" icon={<SearchOutlined />} onClick={queryList}>查询</Button>}<Button icon={<ReloadOutlined />} onClick={resetFilters}>重置</Button></div></div>} actions={<div className="system-action-row">{can("system:role:add") && <Button type="primary" icon={<PlusOutlined />} onClick={() => { roleCreateForm.resetFields(); roleCreateForm.setFieldsValue({ status: 1 }); setRoleCreateOpen(true); }}>新增角色</Button>}</div>}><DataGrid<Role> title="角色列表" preferenceKey="list-columns-system-roles" rowKey="id" loading={loading} dataSource={roles} pagination={listPagination} columns={[
                { title: "角色名称", dataIndex: "name" }, { title: "编码", dataIndex: "code" }, { title: "所属机构", dataIndex: "organizationName", render: (v: string) => <span>{v || "未绑定"}</span> }, { title: "说明", dataIndex: "description" }, { title: "权限数量", dataIndex: "menuIds", render: (v: number[]) => <Tag color="geekblue">{v?.length || 0} 项</Tag> }, { title: "启用状态", dataIndex: "status", width: 110, render: (v: number, item: Role) => <EnableStatusSwitch checked={v === 1} disabled={!can("system:role:edit")} onChange={can("system:role:edit") ? enabled => toggleRole(item, enabled) : undefined}/> }, { title: "创建人", dataIndex: "createdBy", width: 120, ellipsis: true, render: (v: string) => v || "系统管理员" }, { title: "创建时间", dataIndex: "createdAt", width: 170 }, ...((can("system:role:permission") || can("system:role:edit") || can("system:role:delete")) ? [{ title: "操作", width: 220, render: (_: unknown, v: Role) => <Space>{(can("system:role:permission") || can("system:role:edit")) && <Button type="link" icon={<SafetyCertificateOutlined />} onClick={() => { roleEditForm.setFieldsValue(v); setPermissionRole(v); setCheckedMenus(v.isSystem && v.code === "admin" ? allMenuIds : (v.menuIds || [])); setRoleEditOpen(true); }}>编辑与授权</Button>}{can("system:role:delete") && !v.isSystem && <Popconfirm title="确认删除该角色？" onConfirm={() => void remove("roles", v.id)}><Button type="link" danger>删除</Button></Popconfirm>}</Space> }] : [])
            ]}/></ListPageShell>,
        menus: <ListPageShell className="system-list-shell" query={<div className="query-form-row"><label>菜单检索<Input allowClear prefix={<SearchOutlined />} placeholder="菜单名称 / 权限标识 / 路由" value={filters.keyword} onChange={event => setFilters(value => ({ ...value, keyword: event.target.value }))}/></label><label>权限类型<Select allowClear placeholder="请选择类型" value={filters.menuType} onChange={menuType => setFilters(value => ({ ...value, menuType }))} options={[{ value: "directory", label: "目录" }, { value: "menu", label: "菜单" }, { value: "button", label: "按钮" }]}/></label><label>启用状态<Select allowClear placeholder="请选择状态" value={filters.status} onChange={status => setFilters(value => ({ ...value, status }))} options={[{ value: 1, label: "启用" }, { value: 0, label: "停用" }]}/></label><div className="query-actions">{can("system:menu:query") && <Button type="primary" icon={<SearchOutlined />} onClick={queryList}>查询</Button>}<Button icon={<ReloadOutlined />} onClick={resetFilters}>重置</Button></div></div>} actions={<div className="system-action-row">{can("system:menu:add") && <Button type="primary" icon={<PlusOutlined />} onClick={() => { menuForm.resetFields(); menuForm.setFieldsValue({ parentId: 0, menuType: "menu", sortOrder: 0, status: 1 }); setMenuOpen(true); }}>新增权限项</Button>}</div>}><DataGrid<Menu> title="菜单与按钮列表" preferenceKey="list-columns-system-menus" rowKey="id" loading={loading} dataSource={filteredMenuTree} pagination={false} expandable={{ defaultExpandAllRows: true }} columns={[
                { title: "名称", dataIndex: "name", width: 200, ellipsis: { showTitle: false } }, { title: "权限标识", dataIndex: "code", width: 260, ellipsis: { showTitle: false } }, { title: "类型", dataIndex: "menuType", width: 84, render: (v: Menu["menuType"]) => <Tag color={typeMeta[v].color}>{typeMeta[v].text}</Tag> }, { title: "路由路径", dataIndex: "path", width: 220, ellipsis: { showTitle: false }, render: (v: string) => <span className="permission-code">{v || "—"}</span> }, { title: "排序", dataIndex: "sortOrder", width: 72 }, { title: "启用状态", dataIndex: "status", width: 110, render: (value: number, item: Menu) => <EnableStatusSwitch checked={value === 1} disabled={!can("system:menu:edit")} onChange={can("system:menu:edit") ? enabled => toggleMenu(item, enabled) : undefined}/> }, { title: "创建人", dataIndex: "createdBy", width: 120, ellipsis: true, render: (v: string) => v || "系统管理员" }, { title: "创建时间", dataIndex: "createdAt", width: 170 }, ...((can("system:menu:edit") || can("system:menu:delete")) ? [{ title: "操作", width: 140, render: (_: unknown, v: Menu) => <Space>{can("system:menu:edit") && <Button type="link" icon={<EditOutlined />} onClick={() => { menuForm.setFieldsValue(v); setMenuOpen(true); }}>编辑</Button>}{can("system:menu:delete") && <Popconfirm title="确认删除该权限项？" onConfirm={() => void remove("menus", v.id)}><Button type="link" danger>删除</Button></Popconfirm>}</Space> }] : [])
            ]}/></ListPageShell>
    }[page];
    return <section className="management-page">
  {pageContent}
  <Drawer {...drawerProps} title={userForm.getFieldValue("id") ? "编辑用户" : "新增用户"} open={userOpen} onClose={() => setUserOpen(false)} extra={<Space><Button onClick={() => setUserOpen(false)}>取消</Button><Button type="primary" onClick={() => void saveUser()}>保存</Button></Space>}><Form form={userForm} layout="vertical"><Form.Item name="id" hidden><Input /></Form.Item><Form.Item label="用户名" name="username" rules={[{ required: true }]}><Input disabled={!!userForm.getFieldValue("id")}/></Form.Item>{!userForm.getFieldValue("id") && <Form.Item label="初始密码" name="password" rules={[{ required: true, min: 6 }]}><Input.Password /></Form.Item>}<Form.Item label="显示名称" name="displayName" rules={[{ required: true }]}><Input /></Form.Item><Form.Item label="所属角色" name="roleId"><Select allowClear options={roles.filter(r => r.status === 1).map(r => ({ value: r.id, label: r.name }))}/></Form.Item><Form.Item label="所属机构（优先于角色机构）" name="organizationId" extra="未绑定用户机构时，自动采用角色所属机构；两者均未绑定则无业务数据权限。"><Select allowClear showSearch optionFilterProp="label" options={organizations.filter(item => item.status === 1).map(item => ({ value: item.id, label: `${item.name}（${item.code}）` }))}/></Form.Item><Form.Item label="邮箱" name="email"><Input /></Form.Item><Form.Item label="手机号" name="phone"><Input /></Form.Item><Form.Item label="状态" name="status"><Radio.Group options={[{ label: "正常", value: 1 }, { label: "停用", value: 0 }]}/></Form.Item></Form></Drawer>
  <Drawer {...drawerProps} title="新增角色" open={roleCreateOpen} onClose={() => setRoleCreateOpen(false)} extra={<Space><Button onClick={() => setRoleCreateOpen(false)}>取消</Button><Button type="primary" onClick={() => void saveNewRole()}>新增</Button></Space>}><Form form={roleCreateForm} layout="vertical"><Form.Item label="角色名称" name="name" rules={[{ required: true, message: "请输入角色名称" }]}><Input /></Form.Item><Form.Item label="角色编码" name="code" rules={[{ required: true, message: "请输入角色编码" }]}><Input /></Form.Item><Form.Item label="所属机构（数据范围）" name="organizationId" extra="用户未单独绑定机构时，使用此机构及其所有下级机构的数据范围。"><Select allowClear showSearch optionFilterProp="label" options={organizations.filter(item => item.status === 1).map(item => ({ value: item.id, label: `${item.name}（${item.code}）` }))}/></Form.Item><Form.Item label="状态" name="status"><Radio.Group options={[{ label: "正常", value: 1 }, { label: "停用", value: 0 }]}/></Form.Item><Form.Item label="角色说明" name="description"><Input.TextArea rows={6} placeholder="请输入该角色的职责和数据范围说明"/></Form.Item></Form></Drawer>
  <Drawer width="45%" destroyOnHidden title="编辑角色与授权" open={roleEditOpen} onClose={() => { setRoleEditOpen(false); setPermissionRole(null); }} extra={<Space><Button onClick={() => { setRoleEditOpen(false); setPermissionRole(null); }}>取消</Button><Button type="primary" onClick={() => void saveEditedRole()}>保存</Button></Space>}>
   <Form form={roleEditForm} layout="vertical" disabled={!!permissionRole && !can("system:role:edit")}><Form.Item name="id" hidden><Input /></Form.Item><div className="role-basic-grid"><Form.Item label="角色名称" name="name" rules={[{ required: true }]}><Input /></Form.Item><Form.Item label="角色编码" name="code" rules={[{ required: true }]}><Input disabled/></Form.Item><Form.Item label="所属机构（数据范围）" name="organizationId"><Select allowClear showSearch optionFilterProp="label" options={organizations.filter(item => item.status === 1).map(item => ({ value: item.id, label: `${item.name}（${item.code}）` }))}/></Form.Item><Form.Item label="状态" name="status"><Radio.Group options={[{ label: "正常", value: 1 }, { label: "停用", value: 0 }]}/></Form.Item><Form.Item label="角色说明" name="description"><Input.TextArea rows={3}/></Form.Item></div></Form>
   <section className="role-permission-panel"><div className="permission-panel-title"><strong>菜单与按钮权限</strong><span>已选择 {checkedMenus.length} / {allMenuIds.length} 项</span></div>{permissionRole?.isSystem && permissionRole.code === "admin" && <Alert showIcon type="info" message="系统管理员拥有全部已启用权限" description="菜单或按钮停用后，即使管理员已被授权也不会生效。" style={{ marginBottom: 16 }}/>}<div className="permission-toolbar"><Checkbox disabled={!can("system:role:permission") || (permissionRole?.isSystem && permissionRole.code === "admin")} checked={checkedMenus.length === allMenuIds.length && allMenuIds.length > 0} indeterminate={checkedMenus.length > 0 && checkedMenus.length < allMenuIds.length} onChange={e => setCheckedMenus(e.target.checked ? allMenuIds : [])}>全选</Checkbox><span>选择按钮权限时，系统会自动补齐其上级页面和目录权限</span></div><Tree disabled={!can("system:role:permission") || (permissionRole?.isSystem && permissionRole.code === "admin")} checkable defaultExpandAll treeData={permissionTree(menuTree)} checkedKeys={checkedMenus} onCheck={keys => setCheckedMenus(keys as Key[])}/></section>
  </Drawer>
  <Drawer {...drawerProps} title={menuForm.getFieldValue("id") ? "编辑权限项" : "新增权限项"} open={menuOpen} onClose={() => setMenuOpen(false)} extra={<Space><Button onClick={() => setMenuOpen(false)}>取消</Button><Button type="primary" onClick={() => void saveMenu()}>保存</Button></Space>}><Form form={menuForm} layout="vertical"><Form.Item name="id" hidden><Input /></Form.Item><Form.Item label="上级目录或菜单" name="parentId"><Select showSearch optionFilterProp="label" options={[{ value: 0, label: "顶级节点" }, ...menuOptions]}/></Form.Item><Form.Item label="名称" name="name" rules={[{ required: true }]}><Input placeholder="例如：新增用户"/></Form.Item><Form.Item label="权限标识" name="code" rules={[{ required: true }]}><Input disabled={!!menuForm.getFieldValue("id")} placeholder="例如：system:user:add"/></Form.Item><Form.Item label="权限类型" name="menuType"><Radio.Group options={[{ label: "目录", value: "directory" }, { label: "菜单", value: "menu" }, { label: "按钮", value: "button" }]}/></Form.Item><Form.Item label="路由路径" name="path" extra="按钮权限无需填写路由路径"><Input placeholder="/system/users"/></Form.Item><Form.Item label="图标名称" name="icon"><Input placeholder="例如 UserOutlined"/></Form.Item><Form.Item label="显示排序" name="sortOrder"><InputNumber style={{ width: "100%" }}/></Form.Item><Form.Item label="状态" name="status"><Radio.Group options={[{ label: "正常", value: 1 }, { label: "停用", value: 0 }]}/></Form.Item></Form></Drawer>
  <Drawer {...drawerProps} width={480} title={`重置 ${passwordUser?.username || ""} 的密码`} open={!!passwordUser} onClose={() => setPasswordUser(null)} extra={<Space><Button onClick={() => setPasswordUser(null)}>取消</Button><Button type="primary" onClick={async () => { if (!passwordUser)
        return; try {
        const v = await passwordForm.validateFields();
        await axios.put(`/api/admin/users/${passwordUser.id}/password`, v);
        message.success("密码已重置");
        setPasswordUser(null);
    }
    catch (e) {
        if (axios.isAxiosError(e))
            message.error(errorText(e));
    } }}>确认重置</Button></Space>}><Form form={passwordForm} layout="vertical"><Form.Item label="新密码" name="password" rules={[{ required: true, min: 6, message: "密码至少6位" }]}><Input.Password /></Form.Item></Form></Drawer>
 </section>;
}
