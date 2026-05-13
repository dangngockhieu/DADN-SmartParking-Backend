from __future__ import annotations

from collections.abc import Sequence
from typing import Any

import aiomysql

from .config import Settings


class MySQL:
    def __init__(self, pool: aiomysql.Pool) -> None:
        self.pool = pool

    async def fetch_one(self, query: str, args: Sequence[Any] = ()) -> dict[str, Any] | None:
        async with self.pool.acquire() as conn:
            async with conn.cursor(aiomysql.DictCursor) as cur:
                await cur.execute(query, args)
                return await cur.fetchone()

    async def fetch_all(self, query: str, args: Sequence[Any] = ()) -> list[dict[str, Any]]:
        async with self.pool.acquire() as conn:
            async with conn.cursor(aiomysql.DictCursor) as cur:
                await cur.execute(query, args)
                rows = await cur.fetchall()
                return list(rows)

    async def execute(self, query: str, args: Sequence[Any] = ()) -> int:
        async with self.pool.acquire() as conn:
            async with conn.cursor() as cur:
                await cur.execute(query, args)
                await conn.commit()
                return cur.rowcount

    async def close(self) -> None:
        self.pool.close()
        await self.pool.wait_closed()


async def create_mysql(settings: Settings) -> MySQL:
    pool = await aiomysql.create_pool(
        host=settings.db_host,
        port=settings.db_port,
        user=settings.db_user,
        password=settings.db_pass,
        db=settings.db_name,
        minsize=settings.mysql_pool_min_size,
        maxsize=settings.mysql_pool_max_size,
        autocommit=False,
        charset="utf8mb4",
    )
    return MySQL(pool)
