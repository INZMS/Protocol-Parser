import { useSettingsStore } from "../../store/settings";

export default function BrandFooter({ className = "" }: { className?: string }) {
    const settings = useSettingsStore((state) => state.loginPage);
    const details = [
        settings.developerName && `开发商：${settings.developerName}`,
        settings.developerPhone && `联系电话：${settings.developerPhone}`,
        settings.systemVersion && `版本：${settings.systemVersion}`
    ].filter(Boolean);

    return <div className={`brand-footer ${className}`.trim()}>
        <span>© {new Date().getFullYear()} {settings.footerCopyright || settings.systemName}</span>
        {settings.footerSlogan && <><i>·</i><span>{settings.footerSlogan}</span></>}
        {details.map((item) => <span className="brand-footer-detail" key={item}><i>·</i>{item}</span>)}
    </div>;
}
