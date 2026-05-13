from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator

from fastapi import APIRouter, FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware

from .config import settings
from .database import create_mysql
from .gate_service import GateService
from .models import (
    AcceptedResponse,
    CameraPlateRequest,
    CameraPlateResponse,
    RfidScanRequest,
    RfidScanResponse,
    SensorBatchRequest,
)
from .redis_client import create_redis
from .serialization import model_to_dict, to_json


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    redis = create_redis(settings)
    mysql = await create_mysql(settings)
    app.state.redis = redis
    app.state.mysql = mysql
    app.state.gate_service = GateService(mysql, redis, settings)
    try:
        yield
    finally:
        await redis.aclose()
        await mysql.close()


app = FastAPI(title="Smart Parking Async IoT Gateway", version="1.0.0", lifespan=lifespan)

if settings.cors_allowed_origins:
    app.add_middleware(
        CORSMiddleware,
        allow_origins=list(settings.cors_allowed_origins),
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

api = APIRouter(prefix="/api/v1")


@app.get("/health")
async def health(request: Request) -> dict[str, str]:
    await request.app.state.redis.ping()
    await request.app.state.mysql.fetch_one("SELECT 1 AS ok")
    return {"status": "up", "service": "async-iot-gateway"}


@api.post("/iot/sensor/batch", response_model=AcceptedResponse, status_code=202)
async def sensor_batch(request: Request, payload: SensorBatchRequest) -> AcceptedResponse:
    body = model_to_dict(payload)
    stream_id = await request.app.state.redis.xadd(
        settings.sensor_stream,
        {"payload": to_json(body)},
        maxlen=settings.sensor_stream_maxlen,
        approximate=True,
    )
    await request.app.state.redis.hset(
        f"iot:device:{payload.device_mac}:last_batch",
        mapping={
            "sequence": "" if payload.sequence is None else str(payload.sequence),
            "events": str(len(payload.events)),
            "stream_id": stream_id,
        },
    )
    return AcceptedResponse(accepted=True, stream_id=stream_id, events=len(payload.events))


@api.post("/iot/camera", response_model=CameraPlateResponse)
async def camera_plate(request: Request, payload: CameraPlateRequest) -> CameraPlateResponse:
    return await request.app.state.gate_service.handle_camera_plate(payload)


@api.post("/iot/rfid", response_model=RfidScanResponse)
async def rfid_scan(request: Request, payload: RfidScanRequest) -> RfidScanResponse:
    return await request.app.state.gate_service.handle_rfid_scan(payload)


app.include_router(api)
