from __future__ import annotations

import math
from datetime import datetime
from typing import Any

import aiomysql
from redis.asyncio import Redis

from .config import Settings
from .database import MySQL
from .models import CameraPlateRequest, CameraPlateResponse, RfidScanRequest, RfidScanResponse
from .serialization import to_json


def _reject(message: str) -> RfidScanResponse:
    return RfidScanResponse(
        success=False,
        action="reject",
        lcd_line1="REJECTED",
        lcd_line2=message,
        message=message,
    )


def _plate_key(gate_id: int) -> str:
    return f"iot:plate_cache:gate:{gate_id}"


def _gate_lock_key(rfid_uid: str) -> str:
    return f"iot:gate_lock:rfid:{rfid_uid}"


def _calculate_fee(session: dict[str, Any], card: dict[str, Any]) -> int:
    entry_time = session["entry_time"]
    if not isinstance(entry_time, datetime):
        return 5000

    elapsed_seconds = max(0.0, (datetime.now() - entry_time).total_seconds())
    hours = max(1, math.ceil(elapsed_seconds / 3600))
    rate = 5000 if card["card_type"] == "REGISTERED" and card["is_active"] else 7000
    return int(hours * rate)


class GateService:
    def __init__(self, db: MySQL, redis: Redis, settings: Settings) -> None:
        self.db = db
        self.redis = redis
        self.settings = settings

    async def handle_camera_plate(self, req: CameraPlateRequest) -> CameraPlateResponse:
        gate = await self._find_gate(req.gate_id)
        if gate is None:
            return CameraPlateResponse(success=False, message="Gate not found")
        if not gate["is_active"]:
            return CameraPlateResponse(success=False, message="Gate is inactive")

        plate_number = req.plate_number.strip()
        if not plate_number:
            return CameraPlateResponse(success=False, message="Plate number is required")

        await self.redis.set(_plate_key(req.gate_id), plate_number, ex=self.settings.plate_cache_ttl_seconds)
        await self.redis.xadd(
            "iot:gate:audit",
            {"payload": to_json({"type": "CAMERA_PLATE", "gate_id": req.gate_id, "plate_number": plate_number})},
            maxlen=50_000,
            approximate=True,
        )

        return CameraPlateResponse(success=True, message="Plate saved from camera")

    async def handle_rfid_scan(self, req: RfidScanRequest) -> RfidScanResponse:
        lock_key = _gate_lock_key(req.rfid_uid)
        locked = await self.redis.set(lock_key, "1", ex=self.settings.gate_lock_ttl_seconds, nx=True)
        if not locked:
            return _reject("RFID busy")

        try:
            return await self._handle_rfid_scan_locked(req)
        finally:
            await self.redis.delete(lock_key)

    async def _handle_rfid_scan_locked(self, req: RfidScanRequest) -> RfidScanResponse:
        gate = await self._find_gate(req.gate_id)
        if gate is None:
            return _reject("No gate")
        if not gate["is_active"]:
            return _reject("Gate off")
        if gate["mac_address"] != req.mac_address:
            return _reject("MAC mismatch")

        card = await self._find_card(req.rfid_uid)
        if card is None:
            return _reject("Unknown card")
        if not card["is_active"]:
            return _reject("Card disabled")

        plate_number = await self.redis.get(_plate_key(req.gate_id))
        if not plate_number:
            return _reject("No plate")

        if gate["type"] == "ENTRY":
            return await self._handle_entry(gate, card, plate_number)
        if gate["type"] == "EXIT":
            return await self._handle_exit(gate, card, plate_number)
        return _reject("Bad gate type")

    async def _handle_entry(self, gate: dict[str, Any], card: dict[str, Any], plate_number: str) -> RfidScanResponse:
        active_plate = await self.db.fetch_one(
            "SELECT id FROM parking_sessions WHERE plate_number = %s AND is_active = 1 LIMIT 1",
            (plate_number,),
        )
        if active_plate is not None:
            return _reject("Plate in use")

        active_card = await self.db.fetch_one(
            "SELECT id FROM parking_sessions WHERE card_uid = %s AND is_active = 1 LIMIT 1",
            (card["uid"],),
        )
        if active_card is not None:
            return _reject("Card in use")

        available = await self.db.fetch_one(
            "SELECT COUNT(*) AS total FROM parking_slots WHERE lot_id = %s AND status = 'AVAILABLE'",
            (gate["lot_id"],),
        )
        if not available or int(available["total"]) <= 0:
            return _reject("Lot full")

        async with self.db.pool.acquire() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    INSERT INTO parking_sessions
                        (lot_id, card_uid, card_type, plate_number, is_active)
                    VALUES
                        (%s, %s, %s, %s, 1)
                    """,
                    (gate["lot_id"], card["uid"], card["card_type"], plate_number),
                )
                session_id = cur.lastrowid
                await conn.commit()

        await self.redis.delete(_plate_key(gate["id"]))
        await self._write_gate_audit("GATE_ENTRY_ACCEPTED", gate, card, plate_number, session_id)

        return RfidScanResponse(
            success=True,
            action="open_barrier",
            lcd_line1=f"LP:{plate_number}",
            lcd_line2="Welcome!",
            message="Session created",
        )

    async def _handle_exit(self, gate: dict[str, Any], card: dict[str, Any], plate_number: str) -> RfidScanResponse:
        session = await self.db.fetch_one(
            """
            SELECT id, lot_id, card_uid, card_type, plate_number, entry_time
            FROM parking_sessions
            WHERE card_uid = %s AND is_active = 1
            ORDER BY id DESC
            LIMIT 1
            """,
            (card["uid"],),
        )
        if session is None:
            return _reject("No session")
        if session["plate_number"] != plate_number:
            return _reject("Plate mismatch")

        fee = _calculate_fee(session, card)

        if card["user_id"] is None:
            await self._finish_guest_session(session["id"], fee)
            await self.redis.delete(_plate_key(gate["id"]))
            await self._write_gate_audit("GATE_EXIT_CASH_REQUIRED", gate, card, plate_number, session["id"])
            return RfidScanResponse(
                success=True,
                action="reject",
                lcd_line1="Please pay cash",
                lcd_line2=f"Fee:{fee}VND",
                message="Guest must pay cash",
            )

        result = await self._finish_registered_session(session["id"], int(card["user_id"]), fee)
        if result == "INSUFFICIENT_BALANCE":
            return _reject("Insufficient balance")
        if result != "OK":
            return _reject("Payment error")

        await self.redis.delete(_plate_key(gate["id"]))
        await self._write_gate_audit("GATE_EXIT_ACCEPTED", gate, card, plate_number, session["id"])

        return RfidScanResponse(
            success=True,
            action="open_barrier",
            lcd_line1=f"LP:{plate_number}",
            lcd_line2=f"Fee:{fee}VND",
            message="Session finished",
        )

    async def _finish_guest_session(self, session_id: int, fee: int) -> None:
        await self.db.execute(
            """
            UPDATE parking_sessions
            SET exit_time = NOW(), fee = %s, is_active = 0
            WHERE id = %s AND is_active = 1
            """,
            (fee, session_id),
        )

    async def _finish_registered_session(self, session_id: int, user_id: int, fee: int) -> str:
        async with self.db.pool.acquire() as conn:
            try:
                async with conn.cursor(aiomysql.DictCursor) as cur:
                    await cur.execute("START TRANSACTION")
                    await cur.execute("SELECT id, money FROM users WHERE id = %s FOR UPDATE", (user_id,))
                    user = await cur.fetchone()
                    if user is None:
                        await conn.rollback()
                        return "USER_NOT_FOUND"

                    balance_before = int(user["money"])
                    if balance_before < fee:
                        await conn.rollback()
                        return "INSUFFICIENT_BALANCE"

                    balance_after = balance_before - fee
                    await cur.execute(
                        """
                        UPDATE parking_sessions
                        SET exit_time = NOW(), fee = %s, is_active = 0
                        WHERE id = %s AND is_active = 1
                        """,
                        (fee, session_id),
                    )
                    await cur.execute("UPDATE users SET money = %s WHERE id = %s", (balance_after, user_id))
                    await cur.execute(
                        """
                        INSERT INTO wallet_transactions
                            (user_id, `type`, amount, balance_before, balance_after, status, description)
                        VALUES
                            (%s, 'DEDUCT', %s, %s, %s, 'SUCCESS', %s)
                        """,
                        (user_id, fee, balance_before, balance_after, "Parking fee payment"),
                    )
                    await conn.commit()
                    return "OK"
            except Exception:
                await conn.rollback()
                raise

    async def _find_gate(self, gate_id: int) -> dict[str, Any] | None:
        return await self.db.fetch_one(
            """
            SELECT id, type, mac_address, lot_id, is_active
            FROM gates
            WHERE id = %s
            LIMIT 1
            """,
            (gate_id,),
        )

    async def _find_card(self, rfid_uid: str) -> dict[str, Any] | None:
        return await self.db.fetch_one(
            """
            SELECT uid, card_type, user_id, is_active
            FROM rfid_cards
            WHERE uid = %s
            LIMIT 1
            """,
            (rfid_uid,),
        )

    async def _write_gate_audit(
        self,
        event_type: str,
        gate: dict[str, Any],
        card: dict[str, Any],
        plate_number: str,
        session_id: int,
    ) -> None:
        await self.redis.xadd(
            "iot:gate:audit",
            {
                "payload": to_json(
                    {
                        "type": event_type,
                        "gate_id": gate["id"],
                        "lot_id": gate["lot_id"],
                        "card_uid": card["uid"],
                        "plate_number": plate_number,
                        "session_id": session_id,
                    }
                )
            },
            maxlen=50_000,
            approximate=True,
        )
