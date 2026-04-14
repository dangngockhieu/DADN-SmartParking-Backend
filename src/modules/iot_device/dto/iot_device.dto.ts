import { IsNotEmpty, IsOptional, IsString, IsNumber } from 'class-validator';

export class CreateIoTDeviceDTO {
  @IsString()
  @IsNotEmpty()
  macAddress: string;

  @IsString()
  deviceName: string;

  @IsOptional()
  @IsNumber()
  lotId?: number;
}
