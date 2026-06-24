"""
FaceReader simülatör testleri.
Dış bağımlılık yok — stdlib unittest kullanır.
"""

import time
import unittest

from bridge.config import Config
from bridge.face_reader import AttendanceEvent, EventType, FaceReader


class _Cfg:
    """Test için minimal config nesnesi."""
    POLL_INTERVAL_SEC = 1.0
    SIM_MEMBER_COUNT = 10


class TestFaceReaderSimulator(unittest.TestCase):

    def setUp(self):
        self.reader = FaceReader(_Cfg())

    def test_poll_requires_connect(self):
        with self.assertRaises(RuntimeError):
            self.reader.poll()

    def test_connect_disconnect(self):
        self.reader.connect()
        self.assertTrue(self.reader._connected)
        self.reader.disconnect()
        self.assertFalse(self.reader._connected)

    def test_poll_returns_list(self):
        self.reader.connect()
        events = self.reader.poll()
        self.assertIsInstance(events, list)
        self.reader.disconnect()

    def test_events_have_valid_member_ids(self):
        self.reader.connect()
        for _ in range(20):  # birden çok poll ile yeterli örnek al
            for event in self.reader.poll():
                self.assertIsInstance(event, AttendanceEvent)
                self.assertGreaterEqual(event.member_id, 1)
                self.assertLessEqual(event.member_id, _Cfg.SIM_MEMBER_COUNT)
                self.assertIn(event.event_type, list(EventType))
                self.assertAlmostEqual(event.timestamp, time.time(), delta=2.0)
        self.reader.disconnect()

    def test_no_duplicate_entry(self):
        """Aynı üye iki kez giriş yapamaz (çıkmadan)."""
        self.reader.connect()
        seen_entries: set[int] = set()
        for _ in range(30):
            for event in self.reader.poll():
                if event.event_type == EventType.ENTRY:
                    self.assertNotIn(
                        event.member_id, seen_entries,
                        f"Üye {event.member_id} çıkmadan tekrar giriş yaptı"
                    )
                    seen_entries.add(event.member_id)
                elif event.event_type == EventType.EXIT:
                    seen_entries.discard(event.member_id)
        self.reader.disconnect()

    def test_exit_only_for_present_members(self):
        """Salonda olmayan üye çıkış yapamaz."""
        self.reader.connect()
        inside: set[int] = set()
        for _ in range(30):
            for event in self.reader.poll():
                if event.event_type == EventType.ENTRY:
                    inside.add(event.member_id)
                elif event.event_type == EventType.EXIT:
                    self.assertIn(
                        event.member_id, inside,
                        f"Üye {event.member_id} içeride değilken çıkış yaptı"
                    )
                    inside.discard(event.member_id)
        self.reader.disconnect()


class TestConfig(unittest.TestCase):

    def test_defaults(self):
        import os
        os.environ.pop("BACKEND_URL", None)
        cfg = Config()
        self.assertEqual(cfg.BACKEND_URL, "http://localhost:8080")
        self.assertGreater(cfg.POLL_INTERVAL_SEC, 0)
        self.assertGreater(cfg.SIM_MEMBER_COUNT, 0)


if __name__ == "__main__":
    unittest.main()
