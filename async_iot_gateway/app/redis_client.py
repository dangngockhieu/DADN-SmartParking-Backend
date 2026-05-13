from __future__ import annotations

from redis.asyncio import Redis

from .config import Settings


def create_redis(settings: Settings) -> Redis:
    host, _, port = settings.redis_addr.partition(":")
    return Redis(
        host=host or "127.0.0.1",
        port=int(port or "6379"),
        password=settings.redis_password,
        db=settings.redis_db,
        decode_responses=True,
    )
