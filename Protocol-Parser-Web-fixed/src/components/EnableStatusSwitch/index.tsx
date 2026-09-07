import { useState } from "react";
import { Switch } from "antd";

type EnableStatusSwitchProps = {
  checked: boolean;
  disabled?: boolean;
  activeLabel?: string;
  inactiveLabel?: string;
  onChange?: (checked: boolean) => Promise<void> | void;
};

/** A single visual and interaction contract for binary enable/disable status. */
export default function EnableStatusSwitch({
  checked,
  disabled = false,
  activeLabel = "启用",
  inactiveLabel = "停用",
  onChange,
}: EnableStatusSwitchProps) {
  const [loading, setLoading] = useState(false);
  const update = async (next: boolean) => {
    if (!onChange) return;
    setLoading(true);
    try {
      await onChange(next);
    } finally {
      setLoading(false);
    }
  };

  return <Switch className="enable-status-switch" checked={checked} disabled={disabled || !onChange} loading={loading} checkedChildren={activeLabel} unCheckedChildren={inactiveLabel} onChange={next => void update(next)} />;
}
