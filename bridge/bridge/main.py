"""
Köprü servisi giriş noktası.

Çalışma döngüsü:
  1. FaceReader.poll() → olayları al
  2. Her olay için ApiClient.send_event() → backend'e gönder
  3. POLL_INTERVAL_SEC bekle → tekrar başa dön

SIGINT / SIGTERM ile temiz kapatılır.
"""

import logging
import signal
import sys
import time

from .api_client import ApiClient
from .config import Config, load_env
from .face_reader import FaceReader

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%H:%M:%S",
)
logger = logging.getLogger(__name__)

_running = True


def _handle_signal(sig, _frame):
    global _running
    logger.info("Kapatma sinyali alındı (%s), durduruluyor…", signal.Signals(sig).name)
    _running = False


def main() -> None:
    load_env()
    cfg = Config()

    signal.signal(signal.SIGINT, _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    reader = FaceReader(cfg)
    client = ApiClient(cfg)

    logger.info(
        "Köprü servisi başlatılıyor → backend: %s | aralık: %.1fs",
        cfg.BACKEND_URL,
        cfg.POLL_INTERVAL_SEC,
    )
    reader.connect()

    try:
        while _running:
            events = reader.poll()

            if events:
                logger.info("%d olay alındı, gönderiliyor…", len(events))

            for event in events:
                ok = client.send_event(event)
                status = "✓" if ok else "✗"
                logger.info(
                    "%s üye %d → %s",
                    status,
                    event.member_id,
                    event.event_type.value,
                )

            time.sleep(cfg.POLL_INTERVAL_SEC)

    finally:
        reader.disconnect()
        logger.info("Köprü servisi durduruldu.")


if __name__ == "__main__":
    main()
