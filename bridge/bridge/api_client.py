"""
Backend API istemcisi.

Yüz tanıma olaylarını ana sunucuya iletir.
stdlib urllib kullanır — dış bağımlılık yok.
"""

import json
import logging
import urllib.error
import urllib.request
from typing import Any

from .face_reader import AttendanceEvent

logger = logging.getLogger(__name__)

_TIMEOUT_SEC = 5


class ApiClient:
    def __init__(self, config) -> None:
        self._base_url = config.BACKEND_URL.rstrip("/")
        self._secret = config.BRIDGE_SECRET

    def send_event(self, event: AttendanceEvent) -> bool:
        """
        Tek bir yoklama olayını backend'e gönderir.
        Başarıda True, hata durumunda False döner (servis çökmez).
        """
        payload = {
            "member_id": event.member_id,
            "event_type": event.event_type.value,
            "timestamp": event.timestamp,
        }
        return self._post("/api/bridge/attendance", payload)

    # ------------------------------------------------------------------ #
    # İç yardımcılar
    # ------------------------------------------------------------------ #

    def _post(self, path: str, body: dict[str, Any]) -> bool:
        url = self._base_url + path
        data = json.dumps(body).encode()
        req = urllib.request.Request(
            url,
            data=data,
            method="POST",
            headers={
                "Content-Type": "application/json",
                "X-Bridge-Secret": self._secret,
            },
        )
        try:
            with urllib.request.urlopen(req, timeout=_TIMEOUT_SEC) as resp:
                if resp.status == 200:
                    logger.debug("POST %s → 200", path)
                    return True
                logger.warning("POST %s → beklenmedik durum %d", path, resp.status)
                return False
        except urllib.error.HTTPError as exc:
            logger.error("POST %s HTTP hatası: %s %s", path, exc.code, exc.reason)
        except urllib.error.URLError as exc:
            logger.error("POST %s bağlantı hatası: %s", path, exc.reason)
        except TimeoutError:
            logger.error("POST %s zaman aşımı (%ds)", path, _TIMEOUT_SEC)
        return False
