import { ApiProperty } from '@nestjs/swagger';
import { IsNotEmpty, IsString, Matches } from 'class-validator';

export class DateQueryDTO {
  @ApiProperty({
    example: '2026-04-20',
    description: 'Date to query statistics for in YYYY-MM-DD format',
  })
  @IsString()
  @IsNotEmpty({ message: 'Date không được để trống' })
  @Matches(/^\d{4}-\d{2}-\d{2}$/, {
    message: 'Date phải ở định dạng YYYY-MM-DD',
  })
  date: string;
}
