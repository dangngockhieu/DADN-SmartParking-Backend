import { ApiProperty } from "@nestjs/swagger";
import { IsEmail, IsNotEmpty, IsString, Matches, MinLength } from "class-validator";

export class RegisterDTO {
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
    @Matches(/((?=.*\d)|(?=.*\W+))(?![.\n])(?=.*[A-Z])(?=.*[a-z]).*$/, {
        message: 'Mật khẩu phải bao gồm chữ hoa, chữ thường, số và ký tự đặc biệt',
    })
    password: string;

    @ApiProperty({
        example: 'Khieu',
    })
    @IsString()
    @IsNotEmpty({ message: 'FirstName không được để trống' })
    firstName: string;

    @ApiProperty({
        example: 'Dang Ngoc',
    })
    @IsString()
    @IsNotEmpty({ message: 'LastName không được để trống' })
    lastName: string;
}

export class LoginDTO {
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
    password: string;
}

export class ResetPasswordDTO {
    @ApiProperty({
        example: 'test@gmail.com',
    })
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;

    @ApiProperty({
        example: 'abcdefgh',
    })
    @IsString()
    @IsNotEmpty({ message: 'Code không được để trống' })
    code: string;

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