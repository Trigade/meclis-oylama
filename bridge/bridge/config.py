"""
Köprü servisi konfigürasyonu.
Tüm ayarlar ortam değişkeninden (.env) okunur.
"""

import os


def _require(key: str) -> str:
    val = os.environ.get(key)
    if not val:
        raise EnvironmentError(f"Zorunlu ortam değişkeni eksik: {key}")
    return val


def load_env(path: str = ".env") -> None:
    """Basit .env yükleyici — python-dotenv gerektirmez."""
    try:
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, _, value = line.partition("=")
                os.environ.setdefault(key.strip(), value.strip())
    except FileNotFoundError:
        pass  # .env yoksa ortam değişkenlerini doğrudan kullan


class Config:
    # Backend API
    BACKEND_URL: str = os.environ.get("BACKEND_URL", "http://localhost:8080")
    BRIDGE_SECRET: str = os.environ.get("BRIDGE_SECRET", "dev-secret-change-me")

    # Simülatör ayarları
    POLL_INTERVAL_SEC: float = float(os.environ.get("POLL_INTERVAL_SEC", "3.0"))
    SIM_MEMBER_COUNT: int = int(os.environ.get("SIM_MEMBER_COUNT", "32"))

    # TODO: Donanım entegrasyonu — cihaz bağlantı ayarları
    # DEVICE_HOST: str = _require("DEVICE_HOST")
    # DEVICE_PORT: int = int(_require("DEVICE_PORT"))
    # DEVICE_SDK_KEY: str = _require("DEVICE_SDK_KEY")
