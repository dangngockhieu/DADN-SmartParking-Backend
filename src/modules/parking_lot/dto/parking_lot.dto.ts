import { IsNotEmpty, IsOptional, IsString} from "class-validator";

export class CreateParkingLotDTO {
    @IsNotEmpty({ message: 'Email không được để trống' })
    @IsString()
    name: string;

    @IsString()
    @IsNotEmpty({ message: 'Location không được để trống' })
    location: string;
}

export class UpdateParkingLotDTO {
    @IsOptional()
    @IsString()
    @IsNotEmpty()
    name?: string;

    @IsOptional()
    @IsString()
    @IsNotEmpty()
    location?: string;
}
