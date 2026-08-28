import type { ReactNode } from "react";
import { usePermission } from "../../hooks/usePermission";

interface PermissionProps {
    code: string;
    children: ReactNode;
    fallback?: ReactNode;
}

export default function Permission({ code, children, fallback = null }: PermissionProps) {
    return usePermission(code) ? children : fallback;
}
