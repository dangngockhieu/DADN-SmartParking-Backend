import { ApiProperty } from "@nestjs/swagger";
import { SlotStatus } from "@prisma/client";
import { IsEnum, IsNotEmpty, IsNumber, IsString } from "class-validator";

export class AdminUpdateParkingLotDTO {
    @ApiProperty({
        example: 'USER',
    })
    @IsNotEmpty({ message: 'Role không được để trống' })
    @IsEnum(SlotStatus, { message: 'Status phải là AVAILABLE, OCCUPIED hoặc MAINTAIN' })
    status: SlotStatus;
}

export class SensorUpdateParkingLotDTO {
    @ApiProperty({
        example: '00:1B:44:11:3A:B7',
    })
    @IsString()
    @IsNotEmpty({ message: 'Mac không được để trống' })
    mac: string;

    @ApiProperty({
        example: '1',
    })
    @IsNotEmpty({ message: 'Port không được để trống' })
    @IsNumber()
    port: number;

    @ApiProperty({
        example: true,
    })
    @IsNotEmpty({ message: 'IsOccupied không được để trống' })
    isOccupied: boolean;
}

export class CreateParkingSlotDTO {
    @ApiProperty({
        example: 'Vị trí đỗ A',
    })
    @IsNotEmpty({ message: 'Name không được để trống' })
    @IsString()
    name: string;

    @ApiProperty({
        example: '1',
    })
    @IsNotEmpty({ message: 'LotId không được để trống' })
    @IsNumber()
    lotId: number;

    @ApiProperty({
        example: '00:1B:44:11:3A:B7',
    })
    @IsString()
    @IsNotEmpty({ message: 'DeviceMac không được để trống' })
    deviceMac: string;

    @ApiProperty({
        example: '1',
    })
    @IsNumber()
    @IsNotEmpty({ message: 'PortNumber không được để trống' })
    portNumber: number;
}

export class ChangeSlotDeviceDTO {
    @ApiProperty({
        example: '00:1B:44:11:3A:B7',
    })
    @IsString()
    @IsNotEmpty()
    deviceMac: string;

    @ApiProperty({
        example: '1',
    })
    @IsNumber()
    @IsNotEmpty()
    portNumber: number;
}
