from __future__ import annotations

from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class SensorEvent(BaseModel):
    port: int = Field(..., ge=1)
    is_occupied: bool
    changed_at: Optional[datetime] = None


class SensorBatchRequest(BaseModel):
    device_mac: str = Field(..., min_length=1, max_length=50)
    lot_id: Optional[int] = Field(default=None, ge=1)
    sequence: Optional[int] = Field(default=None, ge=0)
    sent_at: Optional[datetime] = None
    events: list[SensorEvent] = Field(..., min_length=1, max_length=256)


class AcceptedResponse(BaseModel):
    accepted: bool
    stream_id: str
    events: int


class CameraPlateRequest(BaseModel):
    gate_id: int = Field(..., ge=1)
    plate_number: str = Field(..., min_length=1, max_length=20)
    captured_at: Optional[datetime] = None


class CameraPlateResponse(BaseModel):
    success: bool
    message: str


class RfidScanRequest(BaseModel):
    gate_id: int = Field(..., ge=1)
    mac_address: str = Field(..., min_length=1, max_length=50)
    rfid_uid: str = Field(..., min_length=1, max_length=20)
    scanned_at: Optional[datetime] = None


class RfidScanResponse(BaseModel):
    success: bool
    action: str
    lcd_line1: str
    lcd_line2: str
    message: str
