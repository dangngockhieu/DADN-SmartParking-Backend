import { IsEmail, IsNotEmpty, IsString, Matches, MinLength } from "class-validator";

export class RegisterDTO {
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;
    @IsString()
    @IsNotEmpty({ message: 'Password không được để trống' })
    @MinLength(6, { message: 'Mật khẩu phải có ít nhất 6 ký tự' })
    @Matches(/((?=.*\d)|(?=.*\W+))(?![.\n])(?=.*[A-Z])(?=.*[a-z]).*$/, {
        message: 'Mật khẩu phải bao gồm chữ hoa, chữ thường, số và ký tự đặc biệt',
    })
    password: string;
    @IsString()
    @IsNotEmpty({ message: 'FirstName không được để trống' })
    firstName: string;
    @IsString()
    @IsNotEmpty({ message: 'LastName không được để trống' })
    lastName: string;
}

export class LoginDTO {
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;
    @IsString()
    @IsNotEmpty({ message: 'Password không được để trống' })
    password: string;
}

export class ResetPasswordDTO {
    @IsEmail({}, { message: 'Email không hợp lệ' })
    @IsNotEmpty({ message: 'Email không được để trống' })
    email: string;
    @IsString()
    @IsNotEmpty({ message: 'Code không được để trống' })
    code: string;
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