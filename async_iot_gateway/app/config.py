from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path

from dotenv import load_dotenv


ROOT_DIR = Path(__file__).resolve().parents[2]


def _load_env() -> None:
    app_env = os.getenv("APP_ENV", "local")
    if app_env == "local":
        load_dotenv(ROOT_DIR / ".env.local")
    load_dotenv(ROOT_DIR / ".env", override=False)


def _get_int(name: str, default: int) -> int:
    raw = os.getenv(name)
    if raw is None or raw == "":
        return default
    return int(raw)


def _get_bool(name: str, default: bool) -> bool:
    raw = os.getenv(name)
    if raw is None or raw == "":
        return default
    return raw.strip().lower() in {"1", "true", "yes", "on"}


@dataclass(frozen=True)
class Settings:
    app_env: str
    gateway_port: int
    db_host: str
    db_port: int
    db_user: str
    db_pass: str
    db_name: str
    mysql_pool_min_size: int
    mysql_pool_max_size: int
    redis_addr: str
    redis_password: str | None
    redis_db: int
    sensor_stream: str
    sensor_group: str
    sensor_consumer: str
    sensor_stream_maxlen: int
    realtime_stream: str
    plate_cache_ttl_seconds: int
    gate_lock_ttl_seconds: int
    worker_batch_window_ms: int
    worker_read_count: int
    worker_block_ms: int
    cors_allowed_origins: tuple[str, ...]
    debug: bool


def load_settings() -> Settings:
    _load_env()

    origins = tuple(
        origin.strip()
        for origin in os.getenv("CORS_ALLOWED_ORIGINS", "").split(",")
        if origin.strip()
    )

    return Settings(
        app_env=os.getenv("APP_ENV", "local"),
        gateway_port=_get_int("IOT_GATEWAY_PORT", 8090),
        db_host=os.getenv("DB_HOST", "127.0.0.1"),
        db_port=_get_int("DB_PORT", 3306),
        db_user=os.getenv("DB_USER", ""),
        db_pass=os.getenv("DB_PASS", ""),
        db_name=os.getenv("DB_NAME", ""),
        mysql_pool_min_size=_get_int("IOT_MYSQL_POOL_MIN_SIZE", 1),
        mysql_pool_max_size=_get_int("IOT_MYSQL_POOL_MAX_SIZE", 10),
        redis_addr=os.getenv("REDIS_ADDR", "127.0.0.1:6379"),
        redis_password=os.getenv("REDIS_PASSWORD") or None,
        redis_db=_get_int("REDIS_DB", 0),
        sensor_stream=os.getenv("IOT_SENSOR_STREAM", "iot:sensor:slot_status"),
        sensor_group=os.getenv("IOT_SENSOR_GROUP", "sensor-workers"),
        sensor_consumer=os.getenv("IOT_SENSOR_CONSUMER", "sensor-worker-1"),
        sensor_stream_maxlen=_get_int("IOT_SENSOR_STREAM_MAXLEN", 100_000),
        realtime_stream=os.getenv("IOT_REALTIME_STREAM", "iot:realtime:parking"),
        plate_cache_ttl_seconds=_get_int("IOT_PLATE_CACHE_TTL_SECONDS", 300),
        gate_lock_ttl_seconds=_get_int("IOT_GATE_LOCK_TTL_SECONDS", 10),
        worker_batch_window_ms=_get_int("IOT_WORKER_BATCH_WINDOW_MS", 1000),
        worker_read_count=_get_int("IOT_WORKER_READ_COUNT", 200),
        worker_block_ms=_get_int("IOT_WORKER_BLOCK_MS", 5000),
        cors_allowed_origins=origins,
        debug=_get_bool("IOT_DEBUG", False),
    )


settings = load_settings()
