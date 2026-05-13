from __future__ import annotations

import asyncio
import json
import logging
from typing import Any

from redis.asyncio import Redis
from redis.exceptions import ResponseError

from .config import settings
from .database import MySQL, create_mysql
from .redis_client import create_redis
from .serialization import to_json


logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("sensor-worker")

SlotKey = tuple[str, int]


async def ensure_group(redis: Redis) -> None:
    try:
        await redis.xgroup_create(settings.sensor_stream, settings.sensor_group, id="0", mkstream=True)
        logger.info("created redis stream group stream=%s group=%s", settings.sensor_stream, settings.sensor_group)
    except ResponseError as exc:
        if "BUSYGROUP" not in str(exc):
            raise


async def read_window(redis: Redis) -> list[tuple[str, dict[str, str]]]:
    response = await redis.xreadgroup(
        settings.sensor_group,
        settings.sensor_consumer,
        {settings.sensor_stream: ">"},
        count=settings.worker_read_count,
        block=settings.worker_block_ms,
    )
    messages = _flatten(response)
    if not messages:
        return []

    deadline = asyncio.get_running_loop().time() + (settings.worker_batch_window_ms / 1000)
    while asyncio.get_running_loop().time() < deadline and len(messages) < settings.worker_read_count:
        extra = await redis.xreadgroup(
            settings.sensor_group,
            settings.sensor_consumer,
            {settings.sensor_stream: ">"},
            count=settings.worker_read_count - len(messages),
            block=1,
        )
        more = _flatten(extra)
        if not more:
            await asyncio.sleep(0.01)
            continue
        messages.extend(more)

    return messages


def _flatten(response: Any) -> list[tuple[str, dict[str, str]]]:
    flattened: list[tuple[str, dict[str, str]]] = []
    for _, stream_messages in response or []:
        for message_id, fields in stream_messages:
            flattened.append((message_id, fields))
    return flattened


def coalesce_messages(messages: list[tuple[str, dict[str, str]]]) -> dict[SlotKey, dict[str, Any]]:
    latest: dict[SlotKey, dict[str, Any]] = {}
    for message_id, fields in messages:
        raw = fields.get("payload")
        if not raw:
            logger.warning("message without payload id=%s", message_id)
            continue
        try:
            batch = json.loads(raw)
        except json.JSONDecodeError:
            logger.warning("invalid json payload id=%s", message_id)
            continue

        device_mac = str(batch.get("device_mac", "")).strip()
        if not device_mac:
            logger.warning("batch without device_mac id=%s", message_id)
            continue

        for event in batch.get("events", []):
            try:
                port = int(event["port"])
            except (KeyError, TypeError, ValueError):
                continue
            latest[(device_mac, port)] = {
                "device_mac": device_mac,
                "port": port,
                "lot_id": batch.get("lot_id"),
                "sequence": batch.get("sequence"),
                "is_occupied": bool(event.get("is_occupied")),
                "changed_at": event.get("changed_at"),
                "source_stream_id": message_id,
            }
    return latest


async def fetch_slots(db: MySQL, keys: list[SlotKey]) -> dict[SlotKey, dict[str, Any]]:
    if not keys:
        return {}

    conditions = " OR ".join(["(device_mac = %s AND port_number = %s)"] * len(keys))
    args: list[Any] = []
    for device_mac, port in keys:
        args.extend([device_mac, port])

    rows = await db.fetch_all(
        f"""
        SELECT id, lot_id, name, device_mac, port_number, status
        FROM parking_slots
        WHERE {conditions}
        """,
        args,
    )
    return {(row["device_mac"], int(row["port_number"])): row for row in rows}


async def process_messages(redis: Redis, db: MySQL, messages: list[tuple[str, dict[str, str]]]) -> int:
    latest = coalesce_messages(messages)
    if not latest:
        return 0

    slots = await fetch_slots(db, list(latest.keys()))
    updates: list[tuple[str, int]] = []
    changed_events: list[dict[str, Any]] = []
    state_updates: list[tuple[str, dict[str, str]]] = []

    for key, event in latest.items():
        slot = slots.get(key)
        if slot is None:
            logger.warning("unknown slot device_mac=%s port=%s", key[0], key[1])
            continue
        if slot["status"] == "MAINTAIN":
            continue

        new_status = "OCCUPIED" if event["is_occupied"] else "AVAILABLE"
        state_updates.append(
            (
                f"iot:slot_state:{slot['lot_id']}:{slot['device_mac']}:{slot['port_number']}",
                {
                "slot_id": str(slot["id"]),
                "lot_id": str(slot["lot_id"]),
                "name": str(slot["name"]),
                "device_mac": str(slot["device_mac"]),
                "port": str(slot["port_number"]),
                "status": new_status,
                "changed_at": "" if event.get("changed_at") is None else str(event["changed_at"]),
                "sequence": "" if event.get("sequence") is None else str(event["sequence"]),
                "source_stream_id": str(event["source_stream_id"]),
                },
            )
        )

        if slot["status"] == new_status:
            continue

        updates.append((new_status, int(slot["id"])))
        changed_events.append(
            {
                "id": slot["id"],
                "lot_id": slot["lot_id"],
                "name": slot["name"],
                "device_mac": slot["device_mac"],
                "port": slot["port_number"],
                "old_status": slot["status"],
                "new_status": new_status,
            }
        )

    if updates:
        async with db.pool.acquire() as conn:
            async with conn.cursor() as cur:
                await cur.executemany("UPDATE parking_slots SET status = %s WHERE id = %s", updates)
                await conn.commit()

    if state_updates:
        pipe = redis.pipeline()
        for key, mapping in state_updates:
            pipe.hset(key, mapping=mapping)
        await pipe.execute()

    if changed_events:
        envelope = {"event": "SLOT_STATUS_CHANGE_BATCH", "data": changed_events}
        await redis.xadd(
            settings.realtime_stream,
            {"payload": to_json(envelope)},
            maxlen=50_000,
            approximate=True,
        )
        await redis.publish("iot:parking:slot_status_changed", to_json(envelope))

    return len(changed_events)


async def run() -> None:
    redis = create_redis(settings)
    db = await create_mysql(settings)
    await ensure_group(redis)
    logger.info(
        "sensor worker started stream=%s group=%s consumer=%s",
        settings.sensor_stream,
        settings.sensor_group,
        settings.sensor_consumer,
    )

    try:
        while True:
            messages = await read_window(redis)
            if not messages:
                continue

            ids = [message_id for message_id, _ in messages]
            try:
                changed = await process_messages(redis, db, messages)
            except Exception:
                logger.exception("failed to process sensor batch; messages will stay pending")
                await asyncio.sleep(1)
                continue

            await redis.xack(settings.sensor_stream, settings.sensor_group, *ids)
            logger.info("processed messages=%d changed_slots=%d", len(messages), changed)
    finally:
        await redis.aclose()
        await db.close()


if __name__ == "__main__":
    asyncio.run(run())
