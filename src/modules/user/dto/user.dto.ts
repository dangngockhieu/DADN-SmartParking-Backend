import { ApiProperty } from "@nestjs/swagger";
import { Role } from "@prisma/client";
import { IsEmail, IsEnum, IsNotEmpty, IsString, Matches, MaxLength, MinLength } from "class-validator";

export class CreateUserDTO {
    @ApiProperty({
        example: 'test@gmail.com',
    })
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;

    @ApiProperty({
        example: 'Password12345@',
    })
    @IsString()
    @IsNotEmpty({ message: 'Password không được để trống' })
    @MinLength(6, { message: 'Mật khẩu phải có ít nhất 6 ký tự' })
    password: string;

    @ApiProperty({
        example: 'John',
    })
    @IsString()
    @IsNotEmpty({ message: 'FirstName không được để trống' })
    firstName: string;

    @ApiProperty({
        example: 'Do',
    })
    @IsString()
    @IsNotEmpty({ message: 'LastName không được để trống' })
    lastName: string;

    @ApiProperty({
        example: 'USER',
    })
    @IsNotEmpty({ message: 'Role không được để trống' })
    @IsEnum(Role, { message: 'Role phải là ADMIN, MANAGER hoặc USER' })
    role: Role;
}

export class ChangePasswordDTO {
    @ApiProperty({
        example: 'OldPassword12345@',
    })
    @IsString()
    @IsNotEmpty({ message: 'Old Password không được để trống' })
    @MinLength(6)
    @MaxLength(30)
    oldPassword: string;

    @ApiProperty({
        example: 'NewPassword12345@',
    })
    @IsString()
    @IsNotEmpty({ message: 'New Password không được để trống' })
    @MinLength(6, { message: 'Mật khẩu phải có ít nhất 6 ký tự' })
    @Matches(/((?=.*\d)|(?=.*\W+))(?![.\n])(?=.*[A-Z])(?=.*[a-z]).*$/, {
        message: 'Mật khẩu phải bao gồm chữ hoa, chữ thường, số và ký tự đặc biệt',
    })
    newPassword: string;

    @ApiProperty({
        example: 'NewPassword12345@',
    })
    @IsString()
    @IsNotEmpty({ message: 'Confirm Password không được để trống' })
    confirmPassword: string;
}

export class ChangeRoleDTO {
    @ApiProperty({
        example: 'USER',
    })
    @IsNotEmpty({ message: 'Role không được để trống' })
    @IsEnum(Role, { message: 'Role phải là ADMIN, MANAGER hoặc USER' })
    newRole: Role;
}