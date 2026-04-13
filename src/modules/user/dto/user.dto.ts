import { Role } from "@prisma/client";
import { IsEmail, IsEnum, IsNotEmpty, IsString, Matches, MaxLength, MinLength } from "class-validator";

export class CreateUserDTO {
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;
    @IsString()
    @IsNotEmpty({ message: 'Password không được để trống' })
    @MinLength(6, { message: 'Mật khẩu phải có ít nhất 6 ký tự' })
    password: string;
    @IsString()
    @IsNotEmpty({ message: 'FirstName không được để trống' })
    firstName: string;
    @IsString()
    @IsNotEmpty({ message: 'LastName không được để trống' })
    lastName: string;
    @IsNotEmpty({ message: 'Role không được để trống' })
    @IsEnum(Role, { message: 'Role phải là ADMIN, MANAGER hoặc USER' })
    role: Role;
}

export class ChangePasswordDTO {
    @IsString()
    @IsNotEmpty({ message: 'Old Password không được để trống' })
    @MinLength(6)
    @MaxLength(30)
    oldPassword: string;
    @IsString()
    @IsNotEmpty({ message: 'New Password không được để trống' })
    @MinLength(6, { message: 'Mật khẩu phải có ít nhất 6 ký tự' })
    @Matches(/((?=.*\d)|(?=.*\W+))(?![.\n])(?=.*[A-Z])(?=.*[a-z]).*$/, {
        message: 'Mật khẩu phải bao gồm chữ hoa, chữ thường, số và ký tự đặc biệt',
    })
    newPassword: string;
    @IsString()
    @IsNotEmpty({ message: 'Confirm Password không được để trống' })
    confirmPassword: string;
}

export class ChangeRoleDTO {
    @IsNotEmpty({ message: 'Role không được để trống' })
    @IsEnum(Role, { message: 'Role phải là ADMIN, MANAGER hoặc USER' })
    newRole: Role;
}