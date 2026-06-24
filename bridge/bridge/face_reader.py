"""
Yüz tanıma soyutlama katmanı.

Şu an: rastgele üye girişi/çıkışı simüle eder.
Donanım geldiğinde: bu dosyayı cihazın SDK'sına bağla.
Geri kalan hiçbir dosya değişmez.

Simülatör davranışı:
  - Her POLL_INTERVAL_SEC saniyede bir 0-2 üye giriş yapar.
  - %20 ihtimalle daha önce giren bir üye çıkış yapar.
  - Üye ID'leri 1..SIM_MEMBER_COUNT arasında rastgele seçilir.
"""

import logging
import random
import time
from dataclasses import dataclass, field
from enum import Enum

logger = logging.getLogger(__name__)


class EventType(str, Enum):
    ENTRY = "entry"
    EXIT = "exit"


@dataclass
class AttendanceEvent:
    member_id: int
    event_type: EventType
    timestamp: float = field(default_factory=time.time)


class FaceReader:
    """
    Yüz tanıma cihazı arayüzü.

    Kullanım:
        reader = FaceReader(config)
        reader.connect()
        for event in reader.poll():
            print(event)
        reader.disconnect()
    """

    def __init__(self, config) -> None:
        self._config = config
        self._connected = False

        # Simülatör durumu — donanımda bu alan olmaz
        self._sim_inside: set[int] = set()

    # ------------------------------------------------------------------ #
    # Bağlantı yönetimi
    # ------------------------------------------------------------------ #

    def connect(self) -> None:
        """
        Cihaza bağlan.

        TODO: Donanım entegrasyonu
            sdk = DeviceSDK(
                host=self._config.DEVICE_HOST,
                port=self._config.DEVICE_PORT,
                key=self._config.DEVICE_SDK_KEY,
            )
            sdk.open()
            self._sdk = sdk
        """
        logger.info("[SIM] Yüz tanıma cihazına bağlanıldı (simülatör modu).")
        self._connected = True

    def disconnect(self) -> None:
        """
        Cihaz bağlantısını kapat.

        TODO: Donanım entegrasyonu
            self._sdk.close()
        """
        logger.info("[SIM] Bağlantı kapatıldı.")
        self._connected = False

    # ------------------------------------------------------------------ #
    # Olay okuma
    # ------------------------------------------------------------------ #

    def poll(self) -> list[AttendanceEvent]:
        """
        Cihazdan birikmiş olayları döndürür.

        TODO: Donanım entegrasyonu
            raw = self._sdk.get_events_since(self._last_cursor)
            self._last_cursor = raw.cursor
            return [self._parse(e) for e in raw.events]

        Şu an simülatör çalışır.
        """
        if not self._connected:
            raise RuntimeError("Cihaza bağlı değil. Önce connect() çağır.")

        return self._simulate_events()

    # ------------------------------------------------------------------ #
    # Simülatör (donanım gelince bu metot silinir)
    # ------------------------------------------------------------------ #

    def _simulate_events(self) -> list[AttendanceEvent]:
        events: list[AttendanceEvent] = []
        max_id = self._config.SIM_MEMBER_COUNT

        # 0-2 yeni giriş
        entry_count = random.randint(0, 2)
        candidates = [i for i in range(1, max_id + 1) if i not in self._sim_inside]
        for mid in random.sample(candidates, min(entry_count, len(candidates))):
            self._sim_inside.add(mid)
            events.append(AttendanceEvent(member_id=mid, event_type=EventType.ENTRY))
            logger.debug("[SIM] Giriş → üye %d", mid)

        # %20 ihtimalle 1 çıkış
        if self._sim_inside and random.random() < 0.20:
            mid = random.choice(list(self._sim_inside))
            self._sim_inside.discard(mid)
            events.append(AttendanceEvent(member_id=mid, event_type=EventType.EXIT))
            logger.debug("[SIM] Çıkış → üye %d", mid)

        return events
