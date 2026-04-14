import { ApiProperty } from "@nestjs/swagger";
import { IsNotEmpty, IsOptional, IsString} from "class-validator";

export class CreateParkingLotDTO {
    @ApiProperty({
        example: 'Bãi đỗ A',
    })
    @IsNotEmpty({ message: 'Name không được để trống' })
    @IsString()
    name: string;

    @ApiProperty({
        example: '123 Đường ABC, Quận XYZ',
    })
    @IsString()
    @IsNotEmpty({ message: 'Location không được để trống' })
    location: string;
}

export class UpdateParkingLotDTO {
    @ApiProperty({
        example: 'Bãi đỗ B',
    })
    @IsOptional()
    @IsString()
    @IsNotEmpty()
    name?: string;

    @ApiProperty({
        example: '456 Đường XYZ, Quận ABC',
    })
    @IsOptional()
    @IsString()
    @IsNotEmpty()
    location?: string;
}
