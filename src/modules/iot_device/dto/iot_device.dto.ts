import { ApiProperty } from '@nestjs/swagger';
import { IsNotEmpty, IsOptional, IsString, IsNumber } from 'class-validator';

export class CreateIoTDeviceDTO {
  @ApiProperty({
    example: 'abcdefghijklmnop',
  })
  @IsString()
  @IsNotEmpty()
  macAddress: string;

  @ApiProperty({
    example: 'Gate Controller A',
  })
  @IsNotEmpty()
  @IsString()
  deviceName: string;

  @ApiProperty({
    example: 1,
  })
  @IsOptional()
  @IsNumber()
  lotId?: number;
}
