import { useMemo } from "react";
import { canAccess, useAuthStore } from "../store/auth";

export function usePermission(code: string) {
    const user = useAuthStore((state) => state.user);
    return useMemo(() => canAccess(user, code), [code, user]);
}

export function usePermissions() {
    const user = useAuthStore((state) => state.user);
    return useMemo(() => new Set(user?.permissions || []), [user]);
}
